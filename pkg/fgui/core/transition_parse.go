package core

import "github.com/chslink/fairygui/pkg/fgui/utils"

func (c *GComponent) setupTransitions(buf *utils.ByteBuffer, start int) {
	if c == nil || buf == nil || start < 0 {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(start, 5) {
		return
	}
	count := int(buf.ReadInt16())
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		if nextPos > buf.Len() {
			nextPos = buf.Len()
		}
		if nextPos <= buf.Pos() {
			buf.SetPos(nextPos)
			continue
		}
		info := TransitionInfo{}
		remaining := func() int { return nextPos - buf.Pos() }
		if remaining() >= 2 {
			if name := buf.ReadS(); name != nil {
				info.Name = *name
			}
		}
		if remaining() >= 4 {
			info.Options = int(buf.ReadInt32())
		}
		if remaining() >= 1 {
			info.AutoPlay = buf.ReadBool()
		}
		if remaining() >= 4 {
			info.AutoPlayTimes = int(buf.ReadInt32())
		}
		if remaining() >= 4 {
			info.AutoPlayDelay = float64(buf.ReadFloat32())
		}
		itemCount := 0
		if remaining() >= 2 {
			itemCount = int(buf.ReadInt16())
		}
		info.Items = make([]TransitionItem, 0, itemCount)
		maxDuration := 0.0
		for j := 0; j < itemCount; j++ {
			if buf.Pos() >= nextPos || nextPos-buf.Pos() < 2 {
				break
			}
			dataLen := int(buf.ReadInt16())
			curPos := buf.Pos()
			if dataLen < 0 || curPos+dataLen > nextPos {
				buf.SetPos(nextPos)
				break
			}
			if parsed := c.parseTransitionItem(buf, curPos, dataLen); parsed != nil {
				end := parsed.Time
				if parsed.Tween != nil {
					end += parsed.Tween.Duration
				}
				if end > maxDuration {
					maxDuration = end
				}
				info.Items = append(info.Items, *parsed)
			}
			buf.SetPos(curPos + dataLen)
		}
		info.ItemCount = len(info.Items)
		info.TotalDuration = maxDuration
		if info.ItemCount > 0 || info.Name != "" {
			c.AddTransition(info)
		}
		buf.SetPos(nextPos)
	}
}

func (c *GComponent) parseTransitionItem(buf *utils.ByteBuffer, start, length int) *TransitionItem {
	if buf == nil || length <= 0 {
		return nil
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	limit := start + length
	if limit > buf.Len() || !buf.Seek(start, 0) {
		return nil
	}
	rem := func() int { return limit - buf.Pos() }
	if rem() <= 0 {
		return nil
	}
	action := transitionActionFromByte(int(buf.ReadByte()))
	item := TransitionItem{Type: action}
	if rem() >= 4 {
		item.Time = float64(buf.ReadFloat32())
	} else {
		return nil
	}
	if rem() >= 2 {
		targetIndex := int(buf.ReadInt16())
		if targetIndex >= 0 {
			item.TargetID = c.resolveTransitionTargetID(targetIndex)
		}
	}
	if rem() >= 2 {
		if label := buf.ReadS(); label != nil {
			item.Label = *label
		}
	}
	hasTween := false
	if rem() > 0 {
		hasTween = buf.ReadBool()
	}
	if hasTween {
		tween := TransitionTween{
			Start: TransitionValue{B1: true, B2: true},
			End:   TransitionValue{B1: true, B2: true},
		}
		if buf.Seek(start, 1) {
			if limit-buf.Pos() >= 4 {
				tween.Duration = float64(buf.ReadFloat32())
			}
			if limit-buf.Pos() >= 1 {
				tween.EaseType = int(buf.ReadByte())
			}
			if limit-buf.Pos() >= 4 {
				tween.Repeat = int(buf.ReadInt32())
			}
			if limit-buf.Pos() >= 1 {
				tween.Yoyo = buf.ReadBool()
			}
			if limit-buf.Pos() >= 2 {
				if endLabel := buf.ReadS(); endLabel != nil {
					tween.EndLabel = *endLabel
				}
			}
		}
		if buf.Seek(start, 2) {
			decodeTransitionValue(buf, limit, action, &tween.Start)
		}
		if buf.Seek(start, 3) {
			decodeTransitionValue(buf, limit, action, &tween.End)
			if buf.Version >= 2 && limit-buf.Pos() >= 4 {
				pathLen := int(buf.ReadInt32())
				if pathLen > 0 {
					points := make([]TransitionPathPoint, 0, pathLen)
					for p := 0; p < pathLen; p++ {
						if limit-buf.Pos() < 1 {
							break
						}
						curveType := int(buf.ReadUint8())
						point := TransitionPathPoint{CurveType: curveType}
						if limit-buf.Pos() < 8 {
							points = append(points, point)
							break
						}
						point.X = float64(buf.ReadFloat32())
						point.Y = float64(buf.ReadFloat32())
						switch curveType {
						case 1:
							if limit-buf.Pos() >= 8 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
							}
						case 2:
							if limit-buf.Pos() >= 16 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
								point.CX2 = float64(buf.ReadFloat32())
								point.CY2 = float64(buf.ReadFloat32())
							} else if limit-buf.Pos() >= 8 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
							}
						}
						points = append(points, point)
					}
					tween.Path = points
				}
			}
		}
		item.Tween = &tween
	} else {
		if buf.Seek(start, 2) {
			decodeTransitionValue(buf, limit, action, &item.Value)
		}
	}
	return &item
}

func transitionActionFromByte(value int) TransitionAction {
	if value < 0 || value > int(TransitionActionUnknown) {
		return TransitionActionUnknown
	}
	return TransitionAction(value)
}

func (c *GComponent) resolveTransitionTargetID(index int) string {
	if c == nil {
		return ""
	}
	child := c.ChildAt(index)
	if child == nil {
		return ""
	}
	if id := child.ResourceID(); id != "" {
		return id
	}
	if name := child.Name(); name != "" {
		return name
	}
	return child.ID()
}

func decodeTransitionValue(buf *utils.ByteBuffer, limit int, action TransitionAction, out *TransitionValue) {
	if out == nil {
		return
	}
	readBool := func() bool {
		if limit-buf.Pos() < 1 {
			return false
		}
		return buf.ReadBool()
	}
	readFloat := func() float64 {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return float64(buf.ReadFloat32())
	}
	readInt := func() int {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return int(buf.ReadInt32())
	}
	readUint32 := func() uint32 {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return buf.ReadUint32()
	}
	readString := func() string {
		if limit-buf.Pos() < 2 {
			return ""
		}
		if s := buf.ReadS(); s != nil {
			return *s
		}
		return ""
	}

	switch action {
	case TransitionActionXY, TransitionActionSize, TransitionActionPivot, TransitionActionSkew:
		out.B1 = readBool()
		out.B2 = readBool()
		out.F1 = readFloat()
		out.F2 = readFloat()
		if buf.Version >= 2 && action == TransitionActionXY && limit-buf.Pos() >= 1 {
			out.B3 = buf.ReadBool()
		}
	case TransitionActionAlpha, TransitionActionRotation:
		out.F1 = readFloat()
	case TransitionActionScale:
		out.F1 = readFloat()
		out.F2 = readFloat()
	case TransitionActionColor:
		out.Color = readUint32()
	case TransitionActionAnimation:
		out.Playing = readBool()
		out.Frame = readInt()
	case TransitionActionVisible:
		out.Visible = readBool()
	case TransitionActionSound:
		out.Sound = readString()
		out.Volume = readFloat()
	case TransitionActionTransition:
		out.TransName = readString()
		out.PlayTimes = readInt()
	case TransitionActionShake:
		out.Amplitude = readFloat()
		out.Duration = readFloat()
	case TransitionActionColorFilter:
		out.F1 = readFloat()
		out.F2 = readFloat()
		out.F3 = readFloat()
		out.F4 = readFloat()
	case TransitionActionText, TransitionActionIcon:
		out.Text = readString()
	}
}
