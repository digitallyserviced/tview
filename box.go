package tview

import (
	"math"

	tcell "github.com/gdamore/tcell/v2"
)

// Box implements Primitive with a background and optional elements such as a
// border and a title. Most subclasses keep their content contained in the box
// but don't necessarily have to.
//
// Note that all classes which subclass from Box will also have access to its
// functions.
//
// See https://github.com/Bios-Marcel/cordless/tview/wiki/Box for an example.
type Box struct {
	// The position of the rect.
	x, y, width, height int

	// The inner rect reserved for the box's content.
	innerX, innerY, innerWidth, innerHeight int

	// Border padding.
	paddingTop, paddingBottom, paddingLeft, paddingRight int

	// Border Size
	borderTop, borderBottom, borderLeft, borderRight bool

	// The box's background color.
	backgroundColor tcell.Color

	// Reverse video.
	reverse bool

	// Whether or not a border is drawn, reducing the box's space for content by
	// two in width and height.
	border        bool
	borderVisible bool
	// The color of the border.
	borderColor tcell.Color

	// The color of the border when the box has focus.
	borderFocusColor tcell.Color

	borderBlinking bool
	borderStyles   *BorderStyle

	// If set to true, the text view will show down and up arrows if there is
	// content out of sight. While box doesn't implement scrolling, this is
	// an abstraction for other components
	indicateOverflow bool

	// The title. Only visible if there is a border, too.
	title string

	// The color of the title.
	titleColor tcell.Color

	// The alignment of the title.
	titleAlign int

	// Provides a way to find out if this box has focus. We always go through
	// this interface because it may be overridden by implementing classes.
	focus Focusable

	// Whether or not this box has focus.
	hasFocus bool

	visible bool

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the primitive's default input handler (nil if
	// nothing should be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	mouseHandler func(event *tcell.EventMouse) bool
	mouseCapture func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse)

	// An optional function which is called before the box is drawn.
	draw func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)
  evented  EventedFunc

	// Handler that gets called when this component receives focus.
	onFocus func()

	// Handler that gets called when this component loses focus.
	onBlur      func()
	borderStyle tcell.Style
	dontClear   bool

	nextFocusableComponents map[FocusDirection][]Primitive
	parent                  Primitive

	onPaste      func([]rune)
	focusManager *FocusManager
	animating    bool
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	b := &Box{
		width:           15,
		height:          10,
		innerX:          -1, // Mark as uninitialized.
		backgroundColor: Styles.PrimitiveBackgroundColor,
		borderStyle: tcell.StyleDefault.Foreground(Styles.BorderColor).
			Background(Styles.PrimitiveBackgroundColor),
		borderColor:             Styles.BorderColor,
		borderFocusColor:        Styles.BorderFocusColor,
		titleColor:              Styles.TitleColor,
		titleAlign:              AlignCenter,
		borderTop:               true,
		borderBottom:            true,
		borderLeft:              true,
		borderRight:             true,
		visible:                 true,
		borderVisible:           true,
		borderStyles:            &Borders,
		animating:               false,
		nextFocusableComponents: make(map[FocusDirection][]Primitive),
	}

	b.focus = b
	return b
}

func (b *Box) GetFocusManager() *FocusManager {
	return b.focusManager
}

func (b *Box) SetFocusManager(fm *FocusManager) *Box {
	return b
}

func (b *Box) SetDontClear(dontClear bool) *Box {
	b.dontClear = dontClear
	return b
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
	return b
}

// SetVisible sets whether the Box should be drawn onto the screen.
func (b *Box) SetVisible(visible bool) {
	b.visible = visible
}

// IsVisible gets whether the Box should be drawn onto the screen.
func (b *Box) IsVisible() bool {
	return b.visible
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) GetRect() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// SetOnFocus sets the handler that gets called when Focus() gets called.
func (b *Box) SetOnFocus(handler func()) {
	b.onFocus = handler
}

// SetOnBlur sets the handler that gets called when Blur() gets called.
func (b *Box) SetOnBlur(handler func()) {
	b.onBlur = handler
}

// GetInnerRect returns the position of the inner rectangle (x, y, width,
// height), without the border and without any padding. Width and height values
// will clamp to 0 and thus never be negative.
func (b *Box) GetInnerRect() (int, int, int, int) {
	if b.innerX >= 0 {
		return b.innerX, b.innerY, b.innerWidth, b.innerHeight
	}
	x, y, width, height := b.GetRect()
	if b.border {
		x += boolToInt(b.borderLeft)
		y += boolToInt(b.borderTop)
		width -= boolToInt(b.borderLeft) + boolToInt(b.borderRight)
		height -= boolToInt(b.borderTop) + boolToInt(b.borderBottom)
	}
	x, y, width, height = x+b.paddingLeft,
		y+b.paddingTop,
		width-b.paddingLeft-b.paddingRight,
		height-b.paddingTop-b.paddingBottom
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return x, y, width, height
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

// SetRect sets a new position of the primitive. Note that this has no effect
// if this primitive is part of a layout (e.g. Flex, Grid) or if it was added
// like this:
//
//	application.SetRoot(b, true)
func (b *Box) Event(f EventerFunc) {
  if b.evented != nil {
    f(b.evented)
  }
}

func (b *Box) SetRect(x, y, width, height int) {
  if x != b.x || y != b.y || width != b.width || height != b.height {
    b.Event(func(f EventedFunc) {
      f("set.rect", b, x,y,width,height)
    })
  }
	b.x = x
	b.y = y
	b.width = width
	b.height = height
	b.innerX = -1 // Mark inner rect as uninitialized.
}

// SetDrawFunc sets a callback function which is invoked after the box primitive
// has been drawn. This allows you to add a more individual style to the box
// (and all primitives which extend it).
//
// The function is provided with the box's dimensions (set via SetRect()). It
// must return the box's inner dimensions (x, y, width, height) which will be
// returned by GetInnerRect(), used by descendent primitives to draw their own
// content.
func (b *Box) SetDrawFunc(
	handler func(screen tcell.Screen, x, y, width, height int) (int, int, int, int),
) *Box {
	b.draw = handler
	return b
}

// GetDrawFunc returns the callback function which was installed with
// SetDrawFunc() or nil if no such function has been installed.
func (b *Box) GetDrawFunc() func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	return b.draw
}
func (b *Box) SetEventedFunc(
	handler EventedFunc,
) *Box {
	b.evented = handler
	return b
}

// GetEventerFunc returns the callback function which was installed with
// SetEventerFunc() or nil if no such function has been installed.
func (b *Box) GetEventedFunc() EventedFunc {
	return b.evented
}

// WrapInputHandler wraps an input handler (see InputHandler()) with the
// functionality to capture input (see SetInputCapture()) before passing it
// on to the provided (default) input handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapInputHandler(
	inputHandler func(*tcell.EventKey, func(p Primitive)),
) func(*tcell.EventKey, func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if b.inputCapture != nil {
			event = b.inputCapture(event)
		}
		if event != nil && inputHandler != nil {
			inputHandler(event, setFocus)
		}

		// return event
	}
}

// InputHandler returns nil.
func (b *Box) InputHandler() func(*tcell.EventKey, func(p Primitive)) {
	return b.WrapInputHandler(nil)
}

// SetInputCapture installs a function which captures key events before they are
// forwarded to the primitive's default key event handler. This function can
// then choose to forward that key event (or a different one) to the default
// handler by returning it. If nil is returned, the default handler will not
// be called.
//
// Providing a nil handler will remove a previously existing handler.
//
// Note that this function will not have an effect on primitives composed of
// other primitives, such as Form, Flex, or Grid. Key events are only captured
// by the primitives that have focus (e.g. InputField) and only one primitive
// can have focus at a time. Composing primitives such as Form pass the focus on
// to their contained primitives and thus never receive any key events
// themselves. Therefore, they cannot intercept key events.
func (b *Box) SetInputCapture(
	capture func(event *tcell.EventKey) *tcell.EventKey,
) *Box {
	b.inputCapture = capture
	return b
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (b *Box) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return b.inputCapture
}

// SetMouseHandler sets the mouse event handler.
func (b *Box) SetMouseHandler(handler func(event *tcell.EventMouse) bool) {
	b.mouseHandler = handler
}

// WrapMouseHandler wraps a mouse event handler (see MouseHandler()) with the
// functionality to capture mouse events (see SetMouseCapture()) before passing
// them on to the provided (default) event handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapMouseHandler(
	mouseHandler func(MouseAction, *tcell.EventMouse, func(p Primitive)) (bool, Primitive),
) func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if b.mouseCapture != nil {
			action, event = b.mouseCapture(action, event)
		}
		if event != nil && mouseHandler != nil {
			consumed, capture = mouseHandler(action, event, setFocus)
		}
		return
	}
}

// MouseHandler returns nil.
func (b *Box) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return b.WrapMouseHandler(
		func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
			if action == MouseLeftClick && b.InRect(event.Position()) {
				setFocus(b)
				consumed = true
			}
			return
		},
	)
}

// SetMouseCapture sets a function which captures mouse events (consisting of
// the original tcell mouse event and the semantic mouse action) before they are
// forwarded to the primitive's default mouse event handler. This function can
// then choose to forward that event (or a different one) by returning it or
// returning a nil mouse event, in which case the default handler will not be
// called.
//
// Providing a nil handler will remove a previously existing handler.
func (b *Box) SetMouseCapture(
	capture func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse),
) *Box {
	b.mouseCapture = capture
	return b
}

// InRect returns true if the given coordinate is within the bounds of the box's
// rectangle.
func (b *Box) InRect(x, y int) bool {
	rectX, rectY, width, height := b.GetRect()
	return x >= rectX && x < rectX+width && y >= rectY && y < rectY+height
}

// GetMouseCapture returns the function installed with SetMouseCapture() or nil
// if no such function has been installed.
func (b *Box) GetMouseCapture() func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse) {
	return b.mouseCapture
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	b.backgroundColor = color
	return b
}

// SetReverse turns on or off the reverse video attribute.
func (b *Box) SetReverse(on bool) *Box {
	b.reverse = on
	return b
}

// SetBorder sets the flag indicating whether or not the box should have a
// border.
func (b *Box) SetBorder(show bool) *Box {
	b.border = show
	return b
}

func (b *Box) SetBorderAttributes(attr tcell.AttrMask) *Box {
	b.borderStyle = b.borderStyle.Attributes(attr)
	return b
}

// GetBorderAttributes returns the border's style attributes.
func (b *Box) GetBorderStyle(bs *BorderStyle) *BorderStyle {
	return b.borderStyles
}

func (b *Box) SetBorderStyle(bs *BorderStyle) *Box {
	b.borderStyles = bs
	return b
}

// GetBorderAttributes returns the border's style attributes.
func (b *Box) GetBorderAttributes() tcell.AttrMask {
	_, _, attr := b.borderStyle.Decompose()
	return attr
}

// SetBorderColor sets the box's border color.
func (b *Box) GetBorderVisible() bool {
	return b.borderVisible
}

func (b *Box) SetBorderVisible(visible bool) *Box {
	b.borderVisible = visible
	return b
}

// SetBorderColor sets the box's border color.
func (b *Box) SetBorderColor(color tcell.Color) *Box {
	b.borderColor = color
	return b
}

// GetBorderColor returns the box's border color.
func (b *Box) GetBorderColor() tcell.Color {
	color, _, _ := b.borderStyle.Decompose()
	return color
}

// GetBackgroundColor returns the box's background color.
func (b *Box) GetBackgroundColor() tcell.Color {
	return b.backgroundColor
}

// SetBorderFocusColor sets the box's border color when focused.
func (b *Box) SetBorderFocusColor(color tcell.Color) *Box {
	b.borderFocusColor = color
	return b
}

// SetBorderSides decides which sides of the border should be shown in case the
// border has been activated.
func (b *Box) SetBorderSides(top, left, bottom, right bool) *Box {
	b.borderTop = top
	b.borderLeft = left
	b.borderBottom = bottom
	b.borderRight = right

	return b
}

// IsBorder indicates whether a border is rendered at all.
func (b *Box) IsBorder() bool {
	return b.border
}

// IsBorderTop indicates whether a border is rendered on the top side of
// this primitive.
func (b *Box) IsBorderTop() bool {
	return b.border && b.borderTop
}

// IsBorderBottom indicates whether a border is rendered on the bottom side
// of this primitive.
func (b *Box) IsBorderBottom() bool {
	return b.border && b.borderBottom
}

// IsBorderRight indicates whether a border is rendered on the right side of
// this primitive.
func (b *Box) IsBorderRight() bool {
	return b.border && b.borderRight
}

// IsBorderLeft indicates whether a border is rendered on the left side of
// this primitive.
func (b *Box) IsBorderLeft() bool {
	return b.border && b.borderLeft
}

func (b *Box) SetBorderBlinking(blinking bool) *Box {
	b.borderBlinking = blinking
	return b
}

// SetTitle sets the box's title.
func (b *Box) SetTitle(title string) *Box {
	b.title = title
	return b
}

// SetTitleColor sets the box's title color.
func (b *Box) SetTitleColor(color tcell.Color) *Box {
	b.titleColor = color
	return b
}

// SetTitleAlign sets the alignment of the title, one of AlignLeft, AlignCenter,
// or AlignRight.
func (b *Box) SetTitleAlign(align int) *Box {
	b.titleAlign = align
	return b
}

// Draw draws this primitive onto the screen.
func (b *Box) Draw(screen tcell.Screen) {
	b.DrawForSubclass(screen, b)
}

// Draw draws this primitive onto the screen.
func (b *Box) DrawForSubclass(screen tcell.Screen, p Primitive) {
	// Don't draw anything if there is no space.
	if b.width <= 0 || b.height <= 0 || !b.visible {
		return
	}

	borderVisible := b.borderVisible

	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(b.backgroundColor).Reverse(b.reverse)
	if !b.dontClear {
		for y := b.y; y < b.y+b.height; y++ {
			for x := b.x; x < b.x+b.width; x++ {
				screen.SetContent(x, y, ' ', nil, background)
			}
		}
	}

	// Draw border.
	b.DrawBorder(borderVisible, background, screen)

	// Call custom draw function.
	if b.draw != nil {
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.draw(
			screen,
			b.x,
			b.y,
			b.width,
			b.height,
		)
	} else {
		// Remember the inner rect.
		b.innerX = -1
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.GetInnerRect()
	}

	if !b.animating {
		// Clamp inner rect to screen.
		width, height := screen.Size()
		if b.innerX < 0 {
			b.innerWidth += b.innerX
			b.innerX = 0
		}
		if b.innerX+b.innerWidth >= width {
			b.innerWidth = width - b.innerX
		}
		if b.innerY+b.innerHeight >= height {
			b.innerHeight = height - b.innerY
		}
		if b.innerY < 0 {
			b.innerHeight += b.innerY
			b.innerY = 0
		}

		if b.innerWidth < 0 {
			b.innerWidth = 0
		}
		if b.innerHeight < 0 {
			b.innerHeight = 0
		}
	}
	return
}

func (b *Box) DrawBorder(borderVisible bool, background tcell.Style, screen tcell.Screen) bool {
	// background = tcell.StyleDefault.Background(0)
	if b.border && b.width >= 2 && b.height >= 1 {
		var borderStyle tcell.Style
		if b.hasFocus {
			if b.borderVisible {
				borderVisible = true
			}
			borderStyle = background.Foreground(b.borderFocusColor)
			IsVtxxx := func() bool {
				return false
			}()
			if IsVtxxx {
				borderStyle = borderStyle.Bold(true)
			}
		} else {
			borderStyle = background.Foreground(b.borderColor)
		}

		if b.borderBlinking {
			borderStyle = borderStyle.Blink(true)
		}

		vertical, horizontal, topLeft, topRight, bottomLeft, bottomRight := ' ', ' ', ' ', ' ', ' ', ' '
		leftVertical, topHorizontal, rightVertical, bottomHorizontal := ' ', ' ', ' ', ' '
		ifc := func(a, b rune) rune {
			if a == rune(0) {
				return b
			}
			return a
		}
		if borderVisible {

			horizontal = b.borderStyles.Horizontal
			vertical = b.borderStyles.Vertical
			topLeft = b.borderStyles.TopLeft
			topRight = b.borderStyles.TopRight
			bottomLeft = b.borderStyles.BottomLeft
			bottomRight = b.borderStyles.BottomRight
			leftVertical = ifc(b.borderStyles.LeftVertical, vertical)
			rightVertical = ifc(b.borderStyles.RightVertical, vertical)
			topHorizontal = ifc(b.borderStyles.TopHorizontal, horizontal)
			bottomHorizontal = ifc(b.borderStyles.BottomHorizontal, horizontal)

		} else {
		}

		if b.borderTop {
			for x := b.x + 1; x < b.x+b.width-1; x++ {
				screen.SetContent(x, b.y, topHorizontal, nil, borderStyle)
			}

			if b.borderLeft {
				screen.SetContent(b.x, b.y, topLeft, nil, borderStyle)
			} else {
				screen.SetContent(b.x, b.y, topHorizontal, nil, borderStyle)
			}

			if b.borderRight {
				screen.SetContent(b.x+b.width-1, b.y, topRight, nil, borderStyle)
			} else {
				screen.SetContent(b.x+b.width-1, b.y, topHorizontal, nil, borderStyle)
			}
		}

		if b.height > 1 {
			if b.borderBottom {
				for x := b.x + 1; x < b.x+b.width-1; x++ {
					screen.SetContent(x, b.y+b.height-1, bottomHorizontal, nil, borderStyle)
				}

				if b.borderLeft {
					screen.SetContent(b.x, b.y+b.height-1, bottomLeft, nil, borderStyle)
				} else {
					screen.SetContent(b.x, b.y+b.height-1, bottomHorizontal, nil, borderStyle)
				}
				if b.borderRight {
					screen.SetContent(
						b.x+b.width-1,
						b.y+b.height-1,
						bottomRight,
						nil,
						borderStyle,
					)
				} else {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, bottomHorizontal, nil, borderStyle)
				}
			}

			if b.borderLeft {
				for y := b.y + 1; y < b.y+b.height-1; y++ {
					screen.SetContent(b.x, y, leftVertical, nil, borderStyle)
				}

				if b.borderTop {
					screen.SetContent(b.x, b.y, topLeft, nil, borderStyle)
				} else {
					screen.SetContent(b.x, b.y, leftVertical, nil, borderStyle)
				}

				if b.borderBottom {
					screen.SetContent(b.x, b.y+b.height-1, bottomLeft, nil, borderStyle)
				} else {
					screen.SetContent(b.x, b.y+b.height-1, leftVertical, nil, borderStyle)
				}
			}

			if b.borderRight {
				for y := b.y + 1; y < b.y+b.height-1; y++ {
					screen.SetContent(b.x+b.width-1, y, rightVertical, nil, borderStyle)
				}

				if b.borderTop {
					screen.SetContent(b.x+b.width-1, b.y, topRight, nil, borderStyle)
				} else {
					screen.SetContent(b.x+b.width-1, b.y, rightVertical, nil, borderStyle)
				}

				if b.borderBottom {
					screen.SetContent(
						b.x+b.width-1,
						b.y+b.height-1,
						bottomRight,
						nil,
						borderStyle,
					)
				} else {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, rightVertical, nil, borderStyle)
				}
			}
		} else if b.height == 1 && !b.borderTop && !b.borderBottom {
			if b.borderLeft {
				screen.SetContent(b.x, b.y, leftVertical, nil, borderStyle)
			}
			if b.borderRight {
				screen.SetContent(b.x+b.width-1, b.y+b.height-1, rightVertical, nil, borderStyle)
			}
		}

		if b.title != "" && b.width >= 4 {
			_, _ = Print(
				screen,
				b.title,
				b.x+1,
				b.y,
				b.width-2,
				b.titleAlign,
				b.titleColor,
			)
			// if len(b.title)-printed > 0 && printed > 0 {
			// 	_, _, style, _ := screen.GetContent(b.x+b.width-2, b.y)
			// 	fg, _, _ := style.Decompose()
			// 	Print(
			// 		screen,
			// 		string(SemigraphicsHorizontalEllipsis),
			// 		b.x+b.width-2,
			// 		b.y,
			// 		1,
			// 		AlignLeft,
			// 		fg,
			// 	)
			// }
		}
	}
	return false
}

// Focus is called when this primitive receives focus.
func (b *Box) Focus(delegate func(p Primitive)) {
	b.hasFocus = true
	if b.onFocus != nil {
		b.onFocus()
	}
}

// Blur is called when this primitive loses focus.
func (b *Box) Blur() {
	b.hasFocus = false
	if b.onBlur != nil {
		b.onBlur()
	}
}

// SetNextFocusableComponents decides which components are to be focused using
// a certain focus direction. If more than one component is passed, the
// priority goes from left-most to right-most. A component will be skipped if
// it is not visible.
func (b *Box) SetNextFocusableComponents(
	direction FocusDirection,
	components ...Primitive,
) {
	b.nextFocusableComponents[direction] = components
}

// NextFocusableComponent decides which component should receive focus next.
// If nil is returned, the focus is retained.
func (b *Box) NextFocusableComponent(direction FocusDirection) Primitive {
	components, avail := b.nextFocusableComponents[direction]
	if avail {
		for _, comp := range components {
			if comp.IsVisible() {
				return comp
			}
		}
	}

	return nil
}

func (b *Box) GetAnimating() bool {
	return b.animating
}

func (b *Box) SetAnimating(anim bool) {
	b.animating = anim
	// return b
}

// HasFocus returns whether or not this primitive has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}

// GetFocusable returns the item's Focusable.
func (b *Box) GetFocusable() Focusable {
	return b.focus
}

// SetIndicateOverflow toggles whether overflow arrows can be drawn in order to
// signal that the component contains content that is out of the viewarea.
func (b *Box) SetIndicateOverflow(indicateOverflow bool) *Box {
	b.indicateOverflow = indicateOverflow
	return b
}

// SetParent defines which component this primitive is currently being
// treated as a child of. This should never be called manually.
func (b *Box) SetParent(parent Primitive) {
	// Reparenting is possible!
	b.parent = parent
}

func (b *Box) DrawOverflow(screen tcell.Screen, showTop, showBottom bool, pct ...float64) {
	if b.indicateOverflow && b.height > 1 {
		overflowIndicatorX := b.innerX + b.innerWidth // - (b.paddingRight)
		style := tcell.StyleDefault.Foreground(Styles.InverseTextColor).
			Background(tcell.GetColor("#202020"))
		bgStyle := tcell.StyleDefault.Background(tcell.GetColor("#202020"))
		topStyle := style
		bottomStyle := style
		pcent := 0.0
		if len(pct) > 0 {
			pcent = pct[0]
		}
		if !showTop {
			topStyle = style.Foreground(tcell.GetColor("#404040"))
		}
		if !showBottom {
			bottomStyle = style.Foreground(tcell.GetColor("#404040"))
		}
		pos := 0.0
		stp := float64(b.innerHeight-1) / 100.0
		if pcent != 0.0 {
			if pcent < 1.0 && pcent > 0.0 {
				pos = float64(pcent*100.0) * stp
			} else if pcent > 1.0 {
				pos = float64(pcent) * stp
			}
			pos = math.Ceil(pos)
		}

		// epsilon := math.Nextafter(1, 2) - 1
		for i := 1; i < b.innerHeight-1; i++ {
			opo := float64(i)
			// fmt.Println(pcent, pos, math.Abs(pos-opo-stp), pos-opo, stp)
			if pos != 0.0 && math.Abs(float64(int(pos-opo))) <= 1 {
				screen.SetContent(
					overflowIndicatorX,
					i+b.innerY+1,
					' ',
					// []rune{' '},
					nil,
					bgStyle.Background(tcell.GetColor("#505050")),
				)
			} else {
				screen.SetContent(
					overflowIndicatorX,
					i+b.innerY+1,
					' ',
					// []rune{' '},
					nil,
					bgStyle,
				)
			}

		}
		// ⇑⇓ ﰵ ﰬ ▲ ⬇⬆ 🭭 🭯 🮦🢁🢃🡹🡻🭩🭫🭭🭯a 🮧▼
		screen.SetContent(
			overflowIndicatorX,
			b.innerY,
			'🭫',
			// []rune{' '},
			nil,
			topStyle.Reverse(true),
		)
		// if showBottom {
			screen.SetContent(
				overflowIndicatorX,
				b.innerY+b.innerHeight+b.paddingBottom-1,
				'🭩',
				// []rune{' '},
				nil,
				bottomStyle.Reverse(true),
			)
		// }
	}
}

// GetParent returns the current parent or nil if the parent hasn't been
// set yet.
func (b *Box) GetParent() Primitive {
	return b.parent
}

// SetOnPaste defines the function that's called in OnPaste.
func (b *Box) SetOnPaste(onPaste func([]rune)) {
	b.onPaste = onPaste
}

// OnPaste is called when a bracketed paste is finished.
func (b *Box) OnPaste(runes []rune) {
	if b.onPaste != nil {
		b.onPaste(runes)
	}
}
