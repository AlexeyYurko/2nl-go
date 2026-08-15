package logic

// NopDelegate implements NLineDelegate with no-op callbacks. Embed it and
// override only the methods you need.
type NopDelegate struct{}

func (NopDelegate) GameDidEnd(*NLine)            {}
func (NopDelegate) GameDidBegin(*NLine)          {}
func (NopDelegate) GameShapeDidLand(*NLine)      {}
func (NopDelegate) GameShapeDidMove(*NLine)      {}
func (NopDelegate) GameShapeDidDrop(*NLine)      {}
func (NopDelegate) ColorShiftMake(*NLine)        {}
func (NopDelegate) GameDidLevelUp(*NLine)        {}
func (NopDelegate) LinesMatched(*NLine, []*Line) {}
