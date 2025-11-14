package render

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// BatchCommand 批处理命令单元
// 借鉴 Unity 版本的 BatchElement 设计
type BatchCommand struct {
	Image      *ebiten.Image  // 渲染图像
	GeoM       ebiten.GeoM    // 几何变换（每帧不同）
	ColorScale ebiten.ColorScale // 颜色缩放
	Blend      ebiten.Blend   // 混合模式
	Filter     ebiten.Filter  // 采样滤镜
	// 批处理标识
	MaterialKey string        // 材质键（用于分组）
}

// RenderCommand 渲染命令类型
type RenderCommand struct {
	Type    CommandType
	Image   *ebiten.Image
	GeoM    ebiten.GeoM
	Options *ebiten.DrawImageOptions
	// 三角形渲染
	Vertices []ebiten.Vertex
	Indices  []uint16
	Texture  *ebiten.Image
}

// CommandType 命令类型
type CommandType int

const (
	CommandTypeImage    CommandType = iota // 图像绘制
	CommandTypeTriangles                   // 三角形绘制
)

// BatchRenderer 批处理渲染器
// 借鉴 Unity 版本的批处理系统设计
type BatchRenderer struct {
	// 按材质分组的命令
	batches map[string][]BatchCommand
	// 独立命令（不适合批处理的）
	standalone []RenderCommand

	// 统计信息
	stats BatchStats
}

// BatchStats 批处理统计
type BatchStats struct {
	BatchCount        int // 批处理组数
	StandaloneCount   int // 独立命令数
	TotalCommands     int // 总命令数
	EstimatedBatches  int // 预估可合并的批次数
}

// NewBatchRenderer 创建批处理器
func NewBatchRenderer() *BatchRenderer {
	return &BatchRenderer{
		batches:     make(map[string][]BatchCommand),
		standalone:  make([]RenderCommand, 0),
		stats:       BatchStats{},
	}
}

// AddCommand 添加渲染命令
func (b *BatchRenderer) AddCommand(img *ebiten.Image, geo ebiten.GeoM, opts *ebiten.DrawImageOptions) {
	if img == nil {
		return
	}

	// 生成材质键（相同材质可以批处理）
	materialKey := b.generateMaterialKey(img, opts)

	// 尝试添加到批处理组
	cmd := BatchCommand{
		Image:      img,
		GeoM:       geo,
		ColorScale: opts.ColorScale,
		Blend:      opts.Blend,
		Filter:     opts.Filter,
		MaterialKey: materialKey,
	}

	b.batches[materialKey] = append(b.batches[materialKey], cmd)
}

// AddTrianglesCommand 添加三角形渲染命令
func (b *BatchRenderer) AddTrianglesCommand(vertices []ebiten.Vertex, indices []uint16, texture *ebiten.Image) {
	if len(vertices) == 0 || len(indices) == 0 {
		return
	}

	cmd := RenderCommand{
		Type:     CommandTypeTriangles,
		Vertices: vertices,
		Indices:  indices,
		Texture:  texture,
	}

	b.standalone = append(b.standalone, cmd)
}

// Flush 批量执行所有命令
// 借鉴 Unity 版本的批处理执行逻辑
func (b *BatchRenderer) Flush(target *ebiten.Image) {
	// 重置统计
	b.stats.BatchCount = 0
	b.stats.StandaloneCount = len(b.standalone)
	b.stats.TotalCommands = 0

	// 执行批处理命令
	for _, cmds := range b.batches {
		if len(cmds) == 0 {
			continue
		}

		b.stats.BatchCount++
		b.stats.TotalCommands += len(cmds)

		// 批处理策略：
		// 1. 相同材质的命令可以合并
		// 2. 但由于 GeoM 每帧不同，仍需分别调用
		// 3. 优势：减少函数调用开销、优化缓存

		for _, cmd := range cmds {
			opts := &ebiten.DrawImageOptions{
				GeoM:       cmd.GeoM,
				ColorScale: cmd.ColorScale,
				Blend:      cmd.Blend,
				Filter:     cmd.Filter,
			}
			target.DrawImage(cmd.Image, opts)
		}
	}

	// 执行独立命令
	for _, cmd := range b.standalone {
		switch cmd.Type {
		case CommandTypeTriangles:
			opts := &ebiten.DrawTrianglesOptions{}
			target.DrawTriangles(cmd.Vertices, cmd.Indices, cmd.Texture, opts)
		}
	}

	// 清理
	b.batches = make(map[string][]BatchCommand)
	b.standalone = b.standalone[:0]
}

// GetStats 获取批处理统计信息
func (b *BatchRenderer) GetStats() BatchStats {
	return b.stats
}

// generateMaterialKey 生成材质键
// 借鉴 Unity 版本的材质键生成策略
func (b *BatchRenderer) generateMaterialKey(img *ebiten.Image, opts *ebiten.DrawImageOptions) string {
	// 使用图像指针和绘制参数组合
	// 注意：GeoM 不参与键生成（每帧都不同）
	return fmt.Sprintf("%p_%v_%v_%v", img, opts.ColorScale, opts.Blend, opts.Filter)
}

// EstimateBatches 预估可优化的批次数
func (b *BatchRenderer) EstimateBatches() int {
	// 估算批处理收益
	// 如果有多个相同材质的命令，可以减少 DrawCall
	batchCount := 0
	for _, cmds := range b.batches {
		if len(cmds) > 1 {
			batchCount++
		}
	}
	b.stats.EstimatedBatches = batchCount
	return batchCount
}

// CommandBuffer 命令缓冲池
// 使用对象池复用命令缓冲区，避免频繁分配
type CommandBuffer struct {
	commands []RenderCommand
}

var commandBufferPool = sync.Pool{
	New: func() interface{} {
		return &CommandBuffer{
			commands: make([]RenderCommand, 0, 64),
		}
	},
}

// GetCommandBuffer 从对象池获取命令缓冲区
func GetCommandBuffer() *CommandBuffer {
	cb := commandBufferPool.Get().(*CommandBuffer)
	cb.commands = cb.commands[:0]
	return cb
}

// PutCommandBuffer 将命令缓冲区返回对象池
func PutCommandBuffer(cb *CommandBuffer) {
	if cb == nil {
		return
	}
	commandBufferPool.Put(cb)
}

// AddCommand 添加命令到缓冲区
func (cb *CommandBuffer) AddCommand(img *ebiten.Image, geo ebiten.GeoM, opts *ebiten.DrawImageOptions) {
	cmd := RenderCommand{
		Type:    CommandTypeImage,
		Image:   img,
		GeoM:    geo,
		Options: opts,
	}
	cb.commands = append(cb.commands, cmd)
}

// AddTrianglesCommand 添加三角形命令
func (cb *CommandBuffer) AddTrianglesCommand(vertices []ebiten.Vertex, indices []uint16, texture *ebiten.Image) {
	cmd := RenderCommand{
		Type:     CommandTypeTriangles,
		Vertices: vertices,
		Indices:  indices,
		Texture:  texture,
	}
	cb.commands = append(cb.commands, cmd)
}

// Execute 执行缓冲区中的所有命令
func (cb *CommandBuffer) Execute(target *ebiten.Image) {
	for _, cmd := range cb.commands {
		switch cmd.Type {
		case CommandTypeImage:
			if cmd.Image != nil && cmd.Options != nil {
				target.DrawImage(cmd.Image, cmd.Options)
			}
		case CommandTypeTriangles:
			if len(cmd.Vertices) > 0 && len(cmd.Indices) > 0 && cmd.Texture != nil {
				opts := &ebiten.DrawTrianglesOptions{}
				target.DrawTriangles(cmd.Vertices, cmd.Indices, cmd.Texture, opts)
			}
		}
	}
}

// GetCommandCount 获取命令数量
func (cb *CommandBuffer) GetCommandCount() int {
	return len(cb.commands)
}
