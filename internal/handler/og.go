package handler

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/JetBrainsMono-Bold.ttf
var jbBoldTTF []byte

//go:embed fonts/JetBrainsMono-Regular.ttf
var jbRegularTTF []byte

//go:embed fonts/JetBrainsMono-SemiBold.ttf
var jbSemiBoldTTF []byte

const (
	x2 = 2 // 2x retina

	ogW = 600 * x2 // 1200px
	ogH = 315 * x2 // 630px
)

// Colors from style.css :root
var (
	cBg      = color.RGBA{R: 0xF7, G: 0xF8, B: 0xFA, A: 255} // --bg
	cWhite   = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 255} // --white
	cBorder  = color.RGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 255} // --gray-200
	cBlue    = color.RGBA{R: 0x5B, G: 0x9B, B: 0xD5, A: 255} // --blue
	cGray900 = color.RGBA{R: 0x1A, G: 0x1A, B: 0x2E, A: 255} // --gray-900
	cGray700 = color.RGBA{R: 0x3A, G: 0x3A, B: 0x4F, A: 255} // --gray-700
	cGray500 = color.RGBA{R: 0x6B, G: 0x72, B: 0x80, A: 255} // --gray-500
	cGray400 = color.RGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 255} // --gray-400
	cGray100 = color.RGBA{R: 0xF3, G: 0xF4, B: 0xF6, A: 255} // --gray-100

	ogRoleCols = map[string][2]color.RGBA{
		"frontend":        {{R: 0xDB, G: 0xEA, B: 0xFE, A: 255}, {R: 0x1D, G: 0x4E, B: 0xD8, A: 255}},
		"backend":         {{R: 0xD1, G: 0xFA, B: 0xE5, A: 255}, {R: 0x06, G: 0x5F, B: 0x46, A: 255}},
		"fullstack":       {{R: 0xED, G: 0xE9, B: 0xFE, A: 255}, {R: 0x5B, G: 0x21, B: 0xB6, A: 255}},
		"project-manager": {{R: 0xFE, G: 0xF3, B: 0xC7, A: 255}, {R: 0x92, G: 0x40, B: 0x0E, A: 255}},
		"product-manager": {{R: 0xFE, G: 0xE2, B: 0xE2, A: 255}, {R: 0x99, G: 0x1B, B: 0x1B, A: 255}},
		"ux-ui-designer":  {{R: 0xFC, G: 0xE7, B: 0xF3, A: 255}, {R: 0x9D, G: 0x17, B: 0x4D, A: 255}},
		"analyst":         {{R: 0xCF, G: 0xFA, B: 0xFE, A: 255}, {R: 0x15, G: 0x5E, B: 0x75, A: 255}},
		"logo-designer":   {{R: 0xFF, G: 0xED, B: 0xD5, A: 255}, {R: 0x9A, G: 0x34, B: 0x12, A: 255}},
		"qa":              {{R: 0xE0, G: 0xE7, B: 0xFF, A: 255}, {R: 0x37, G: 0x30, B: 0xA3, A: 255}},
		"devops":          {{R: 0xCC, G: 0xFB, B: 0xF1, A: 255}, {R: 0x11, G: 0x5E, B: 0x59, A: 255}},
		"ios":             {{R: 0xF3, G: 0xE8, B: 0xFF, A: 255}, {R: 0x6B, G: 0x21, B: 0xA8, A: 255}},
		"android":         {{R: 0xDC, G: 0xFC, B: 0xE7, A: 255}, {R: 0x16, G: 0x65, B: 0x34, A: 255}},
		"flutter":         {{R: 0xE0, G: 0xF2, B: 0xFE, A: 255}, {R: 0x07, G: 0x59, B: 0x85, A: 255}},
	}
)

type ogFonts struct {
	logo, title, body, badge, small font.Face
}

var (
	ogFontsOnce     sync.Once
	ogFontsInstance *ogFonts
	ogFontsErr      error
)

func getOGFonts() (*ogFonts, error) {
	ogFontsOnce.Do(func() {
		bold, err := opentype.Parse(jbBoldTTF)
		if err != nil {
			ogFontsErr = err
			return
		}
		regular, err := opentype.Parse(jbRegularTTF)
		if err != nil {
			ogFontsErr = err
			return
		}
		semibold, err := opentype.Parse(jbSemiBoldTTF)
		if err != nil {
			ogFontsErr = err
			return
		}

		mk := func(f *opentype.Font, sz float64) font.Face {
			fc, _ := opentype.NewFace(f, &opentype.FaceOptions{
				Size: sz * x2, DPI: 72, Hinting: font.HintingFull,
			})
			return fc
		}

		ogFontsInstance = &ogFonts{
			logo:  mk(bold, 13),    // logo: 700 weight
			title: mk(bold, 20),    // project title: 700 weight, ~1.5rem
			body:  mk(regular, 11), // description: 400 weight, ~0.85rem
			badge: mk(semibold, 9), // badges: 600 weight, ~0.7rem
			small: mk(regular, 8),  // small text: 400, ~0.7rem
		}
	})
	return ogFontsInstance, ogFontsErr
}

func (h *Handler) handleOGImage(w http.ResponseWriter, r *http.Request) {
	project := h.projectBySlug(w, r)
	if project == nil {
		return
	}
	if project.Status != "active" {
		http.NotFound(w, r)
		return
	}
	fonts, err := getOGFonts()
	if err != nil {
		log.Printf("og: font error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, ogW, ogH))

	margin := 16 * x2  // space around card
	pad := 24 * x2     // padding inside card
	radius := 12 * x2  // --radius-lg: 12px
	badgeR := 4 * x2   // badge radius: 4px

	// Background (#F7F8FA)
	ogFill(img, 0, 0, ogW, ogH, cBg)

	// White card with 1px border, rounded corners
	cardX, cardY := margin, margin
	cardW, cardH := ogW-2*margin, ogH-2*margin
	ogRRect(img, cardX-1, cardY-1, cardW+2, cardH+2, radius+1, cBorder)
	ogRRect(img, cardX, cardY, cardW, cardH, radius, cWhite)

	cx := cardX + pad
	cw := cardW - 2*pad
	y := cardY + pad

	// Logo: "svyaz" (gray-900) + "_" (blue)
	ogTxt(img, fonts.logo, cGray900, cx, y, "svyaz")
	lw := ogMsr(fonts.logo, "svyaz")
	ogTxt(img, fonts.logo, cBlue, cx+lw, y, "_")
	y += 24 * x2

	// Title (bold, gray-900, word-wrapped, max 2 lines)
	tl := ogWrapLines(fonts.title, project.Title, cw)
	if len(tl) > 2 {
		tl = tl[:2]
		rn := []rune(tl[1])
		if len(rn) > 3 {
			tl[1] = string(rn[:len(rn)-3]) + "..."
		}
	}
	for _, line := range tl {
		ogTxt(img, fonts.title, cGray900, cx, y, line)
		y += 26 * x2
	}

	// Description (regular, gray-500, max 2 lines)
	if project.Description != "" {
		y += 4 * x2
		desc := strings.ReplaceAll(project.Description, "\n", " ")
		rn := []rune(desc)
		if len(rn) > 200 {
			desc = string(rn[:200]) + "..."
		}
		dl := ogWrapLines(fonts.body, desc, cw)
		if len(dl) > 2 {
			dl = dl[:2]
			lr := []rune(dl[1])
			if len(lr) > 3 {
				dl[1] = string(lr[:len(lr)-3]) + "..."
			}
		}
		for _, line := range dl {
			ogTxt(img, fonts.body, cGray500, cx, y, line)
			y += 15 * x2
		}
	}

	// Role badges (semibold, colored bg, rounded)
	if len(project.Roles) > 0 {
		y += 10 * x2
		bx := cx
		hp, vp := 7*x2, 3*x2
		met := fonts.badge.Metrics()
		bh := met.Ascent.Ceil() + met.Descent.Ceil() + vp*2
		for _, role := range project.Roles {
			cols, ok := ogRoleCols[role.Slug]
			if !ok {
				cols = [2]color.RGBA{cGray100, cGray700}
			}
			tw := ogMsr(fonts.badge, role.Name)
			bw := tw + hp*2
			if bx+bw > cx+cw {
				break
			}
			ogRRect(img, bx, y, bw, bh, badgeR, cols[0])
			ogTxt(img, fonts.badge, cols[1], bx+hp, y+vp, role.Name)
			bx += bw + 6*x2
		}
		y += bh + 8*x2
	}

	// Stack tags (regular, gray-100 bg, gray-700 text)
	if len(project.Stack) > 0 {
		y += 2 * x2
		bx := cx
		hp, vp := 5*x2, 2*x2
		met := fonts.small.Metrics()
		bh := met.Ascent.Ceil() + met.Descent.Ceil() + vp*2
		for _, tag := range project.Stack {
			tw := ogMsr(fonts.small, tag)
			bw := tw + hp*2
			if bx+bw > cx+cw {
				break
			}
			ogRRect(img, bx, y, bw, bh, 3*x2, cGray100)
			ogTxt(img, fonts.small, cGray700, bx+hp, y+vp, tag)
			bx += bw + 4*x2
		}
	}

	// Footer: separator line + URL
	fy := cardY + cardH - pad - 10*x2
	ogFill(img, cx, fy, cw, 1, cBorder)
	ogTxt(img, fonts.small, cGray400, cx, fy+6*x2, "svyaz.fitra.tech")

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if err := png.Encode(w, img); err != nil {
		log.Printf("og: encode error: %v", err)
	}
}

// --- Drawing primitives ---

func ogFill(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), &image.Uniform{c}, image.Point{}, draw.Src)
}

func ogRRect(img *image.RGBA, x, y, w, h, r int, c color.RGBA) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if ogInRR(dx, dy, w, h, r) {
				img.Set(x+dx, y+dy, c)
			}
		}
	}
}

func ogInRR(dx, dy, w, h, r int) bool {
	if dx < 0 || dy < 0 || dx >= w || dy >= h {
		return false
	}
	corners := [][2]int{{r, r}, {w - r - 1, r}, {r, h - r - 1}, {w - r - 1, h - r - 1}}
	for _, co := range corners {
		cx, cy := co[0], co[1]
		inX := (dx <= cx && cx == r) || (dx >= cx && cx == w-r-1)
		inY := (dy <= cy && cy == r) || (dy >= cy && cy == h-r-1)
		if inX && inY {
			fx, fy := float64(dx-cx), float64(dy-cy)
			if fx*fx+fy*fy > float64(r*r) {
				return false
			}
		}
	}
	return true
}

func ogTxt(img *image.RGBA, face font.Face, col color.Color, x, y int, text string) {
	(&font.Drawer{
		Dst: img, Src: &image.Uniform{col}, Face: face,
		Dot: fixed.P(x, y+face.Metrics().Ascent.Ceil()),
	}).DrawString(text)
}

func ogMsr(face font.Face, text string) int {
	return (&font.Drawer{Face: face}).MeasureString(text).Ceil()
}

func ogWrapLines(face font.Face, text string, maxW int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if t := cur + " " + w; ogMsr(face, t) <= maxW {
			cur = t
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	return append(lines, cur)
}
