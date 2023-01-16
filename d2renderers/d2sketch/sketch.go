package d2sketch

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	_ "embed"

	"github.com/dop251/goja"

	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/color"
	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/label"
	"oss.terrastruct.com/d2/lib/svg"
	svg_style "oss.terrastruct.com/d2/lib/svg/style"
	"oss.terrastruct.com/util-go/go2"
)

//go:embed rough.js
var roughJS string

//go:embed setup.js
var setupJS string

type Runner goja.Runtime

var baseRoughProps = `fillWeight: 2.0,
hachureGap: 16,
fillStyle: "solid",
bowing: 2,
seed: 1,`

var floatRE = regexp.MustCompile(`(\d+)\.(\d+)`)

func (r *Runner) run(js string) (goja.Value, error) {
	vm := (*goja.Runtime)(r)
	return vm.RunString(js)
}

func InitSketchVM() (*Runner, error) {
	vm := goja.New()
	if _, err := vm.RunString(roughJS); err != nil {
		return nil, err
	}
	if _, err := vm.RunString(setupJS); err != nil {
		return nil, err
	}
	r := Runner(*vm)
	return &r, nil
}

// DefineFillPatterns adds reusable patterns that are overlayed on shapes with
// fill. This gives it a subtle streaky effect that subtly looks hand-drawn but
// not distractingly so.
func DefineFillPatterns() string {
	out := "<defs>"
	out += defineFillPattern("bright", "rgba(0, 0, 0, 0.1)")
	out += defineFillPattern("normal", "rgba(0, 0, 0, 0.16)")
	out += defineFillPattern("dark", "rgba(0, 0, 0, 0.32)")
	out += defineFillPattern("darker", "rgba(255, 255, 255, 0.24)")
	out += "</defs>"
	return out
}

func defineFillPattern(luminanceCategory, fill string) string {
	return fmt.Sprintf(`<pattern id="streaks-%s" x="0" y="0" width="100" height="100" patternUnits="userSpaceOnUse">
		<path fill="%s" fill-rule="evenodd" clip-rule="evenodd" d="M58.1193 0H58.1703L55.4939 2.67644L58.1193 0ZM45.7725 0H45.811L41.2851 4.61498L42.7191 3.29325L37.0824 8.92997L35.0554 10.9569L32.0719 13.9404L29.6229 16.5017L27.1738 19.0631L25.8089 20.2034L23.2195 22.6244L18.181 27.6068L23.8178 21.97L27.0615 18.9508L33.8666 11.9773L33.1562 12.5194L37.0262 8.87383L40.784 5.11602L38.0299 7.64561L45.7725 0ZM23.1079 0H23.108L21.5814 1.66688L20.3126 2.79534L23.1079 0ZM7.53869 0H7.54254L7.50005 0.035944L7.53869 0ZM2.49995 0H2.52362L0.900245 1.59971L2.49995 0ZM0 3.64398V3.60744L0.278386 3.36559L0 3.64398ZM0 18.6564V18.5398L0.67985 17.8416L3.4459 15.0755L1.15701 17.1333L2.78713 15.6022L6.01437 12.507L8.5168 9.87253L5.15803 13.2313L11.0357 7.25453L10.4926 7.89678L13.6868 4.7686L8.54982 9.90555L7.05177 11.5687L4.68087 13.9396L0.729379 17.8911L3.01827 15.8333L0 18.6564ZM0 69.2431V69.178L1.64651 67.4763L1.46347 67.7796L5.84063 63.4025L4.42167 64.9016L0 69.4007V69.3408L0.247596 68.9955L0 69.2431ZM2.51594 100H2.49238L5.19989 97.2925L7.70071 95.0162L12.8713 89.6772L12.3094 90.0707L15.288 87.3167L18.1542 84.4504L16.0269 86.3532L22.8752 79.6172L18.5364 84.0683L19.6435 83.0734L15.3441 87.3728L13.798 88.9189L11.5224 91.1945L9.66768 93.1615L7.81297 95.1285L6.74529 95.9716L4.75024 97.7983L2.51594 100ZM7.54255 100H7.5387L9.81396 97.884L8.46606 99.2189L7.54255 100ZM45.8189 100H45.7807L46.9912 98.8047L45.8189 100ZM58.1784 100H58.1272L62.2952 95.7511L66.1408 91.9055L63.0037 94.8115L65.2507 92.6635L69.7117 88.3346L73.2165 84.6977L68.5469 89.3673L76.7379 81.0773L75.9634 81.9509L80.3913 77.5889L73.2496 84.7307L71.1346 87.0107L67.8384 90.3069L62.3447 95.8006L65.4818 92.8947L61.2625 96.9159L58.1784 100ZM75.4277 100H75.229L82.1834 92.9039L81.3403 93.5787L86.0063 89.1371L90.5601 84.5833L87.2464 87.6725L98.0937 76.9375L91.1673 83.9761L92.8932 82.3625L86.0625 89.1933L83.6062 91.6496L79.9907 95.265L77.011 98.357L75.4277 100ZM100 18.5398V18.6563L99.9556 18.6979L95.8065 22.847L100 18.5398ZM100 3.60743V3.64398L99.6791 3.9649L99.2094 4.29428L100 3.60743ZM75.4201 0L74.0312 1.4412L72.401 2.84687L69.281 5.79854L63.1812 11.8422L70.0119 5.01151L73.919 1.32893L75.2214 0H75.4201ZM100 69.1858V69.2509L98.059 71.1919L100 69.1858ZM100 69.3486V69.4085L99.8414 69.5698L100 69.3486ZM41.9398 28.8254L53.6223 16.993L52.5215 18.2437L54.7428 16.0575L54.6875 16.0759L54.8008 16.0004L58.842 12.0231L54.9925 15.8726L55.1085 15.7953L54.898 16.0058L54.84 16.0251L48.6523 22.2128L45.6419 25.473L40.9389 30.1759L33.1007 38.0142L37.5866 33.878L31.558 39.6068L23.3278 47.837L33.0257 37.9393L38.5125 32.4525L34.0266 36.5887L37.2369 33.5283L43.6074 27.3576L48.6023 22.1628L41.9398 28.8254ZM41.0977 17.0531L39.718 18.2925L40.312 17.8388L41.0977 17.0531ZM36.875 20.3106L48.1601 7.88137L42.3438 13.7478L36.875 20.3106ZM35.7125 25.8109L34.3328 27.0503L34.9268 26.5966L35.7125 25.8109ZM17.7022 39.7534L19.0819 38.514L18.8092 38.7867L36.7575 21.8045L23.1569 35.3051L13.5771 43.7372L18.1448 39.4154L17.7022 39.7534ZM3.48102 28.9281L1.53562 30.8735L1.22228 31.0465L0.0765686 32.3326L1.60579 30.9437L2.57849 29.971L3.48102 28.9281ZM0.953463 26.2027L19.5702 7.58594L9.31575 18.6078L0.953463 26.2027ZM23.7175 12.11L17.9339 18.0875L21.4622 14.5592L20.8074 15.4725L28.1915 7.95918L30.4791 5.54232L23.4224 12.599L23.7175 12.11ZM43.4641 43.1538L40.7872 46.1552L42.4907 44.4517L42.3285 45.0465L45.8166 41.3421L46.8441 40.0983L43.4371 43.5053L43.4641 43.1538ZM1.32715 48.3271L8.0918 41.5625L4.3657 45.5674L1.32715 48.3271ZM11.1479 31.2556L11.5689 30.975L11.3584 31.1855L11.1479 31.2556ZM11.9898 27.4667L12.2003 27.2562L11.7793 27.5369L11.9898 27.4667ZM11.3585 34.5531L11.148 34.7636L10.9375 34.8338L11.3585 34.5531ZM72.929 28.5457L82.2965 19.0792L81.4043 20.0705L86.4597 15.0811L78.2983 23.2425L75.8697 25.8362L72.1029 29.603L65.8249 35.881L69.3934 32.5437L64.5858 37.1531L57.994 43.745L65.7754 35.8314L70.17 31.4369L66.6015 34.7742L69.1623 32.3125L74.2507 27.3562L78.2653 23.2095L72.929 28.5457ZM82.6674 1.83549L84.3245 0.31872L83.3724 1.27088L82.6674 1.83549ZM64.5872 16.1312L62.9301 17.648L63.6351 17.0834L64.5872 16.1312ZM70.868 9.85044L80.0048 1.1214L74.6221 6.47142L70.868 9.85044ZM90.2409 41.9448L70.7578 61.4279L79.5093 53.4795L90.2409 41.9448ZM91.8088 42.5434L95.3963 38.8357L95.2132 39.139L99.5904 34.7618L98.1714 36.261L93.5912 40.9214L93.9973 40.3549L91.8088 42.5434ZM94.331 12.8233L89.9853 17.1691L89.2853 17.5555L86.7259 20.4284L90.142 17.3258L92.3149 15.1529L94.331 12.8233ZM44.7972 62.3259L76.9824 30.1406L59.2542 49.1955L44.7972 62.3259ZM77.1482 40.321L70.1709 47.5323L70 47.6463L70.0895 47.6164L68.1916 49.5779L70.185 47.5846L70.2105 47.5761L70.421 47.3656L70.37 47.3996L73.6557 44.1139L72.6416 45.5283L84.0768 33.893L87.6194 30.1502L76.6913 41.0783L77.1482 40.321ZM50.5355 34.3137L72.6617 12.1875L60.4955 25.3084L50.5355 34.3137ZM70.2104 44.0681L70.6314 43.7875L70.4209 43.998L70.2104 44.0681ZM71.263 40.0687L70.842 40.3494L71.0525 40.2792L71.263 40.0687ZM55.1084 12.4355L55.3189 12.225L54.8979 12.5056L55.1084 12.4355ZM48.8718 15.5785L60.2075 4.70496L49.4056 15.4006L48.8718 15.5785ZM23.7636 57.4491L29.9099 51.5854L26.1656 55.6123L27.2361 54.8244L23.435 58.6255L22.0681 59.9924L20.0562 62.0042L18.5082 63.8349L16.9601 65.6656L15.8328 66.2277L13.9315 67.7051L10.4821 71.0132L14.2832 67.2121L16.6775 65.383L21.1113 60.5253L20.477 60.7357L23.2937 58.4842L25.8277 55.9502L23.7636 57.4491ZM48.3825 74.1824L44.8832 77.8523L46.9145 75.8211L45.4748 77.4881L43.4493 79.2862L42.4082 80.1568L43.9215 79.0414L42.2487 80.7143L39.3752 83.8151L41.8844 81.3059L43.8473 79.6842L42.334 80.7995L44.7237 78.4098L46.1576 76.976L46.9713 75.8779L50.078 72.7713L48.1093 74.6262L48.3825 74.1824ZM29.2877 62.9906L29.0772 63.2011L28.8667 63.2713L29.2877 62.9906ZM29.7088 59.4823L29.9193 59.2719L29.4983 59.5525L29.7088 59.4823ZM29.0772 66.5687L28.8667 66.7792L28.6562 66.8494L29.0772 66.5687ZM22.9729 68.748L23.1834 68.5375L22.7624 68.8181L22.9729 68.748ZM3.8147e-05 91.7593L13.2499 79.1355L6.5001 86.2595L3.8147e-05 91.7593ZM16.0685 87.9974L17.1375 87.0687L16.5382 87.668L16.0685 87.9974ZM21.7869 79.3344L20.7179 80.263L21.1876 79.9337L21.7869 79.3344ZM12.3607 95.0755L13.4298 94.1469L12.8304 94.7462L12.3607 95.0755ZM42.7176 59.3801L43.2789 58.8187L43.0684 59.1696L42.7877 59.4502L42.2966 59.801L42.5772 59.3801H42.7176ZM26.3124 49.3152L24.3599 51.2676L23.996 51.3918L22.8956 52.732L24.4798 51.3875L25.456 50.4113L26.3124 49.3152ZM39.0689 63.3097L38.5777 63.6606L39.56 62.6782L39.0689 63.3097ZM20.3574 55.8032L19.3751 56.7856L19.8662 56.4347L20.3574 55.8032ZM39.9297 64.195L41.5504 62.3779L41.534 62.5907L43.5967 60.528L42.9746 61.2811L40.8628 63.5238L40.961 63.1637L39.9297 64.195ZM22.3921 55.457L21.3998 56.5696L22.0313 55.9381L21.9711 56.1587L23.2642 54.7854L23.6451 54.3243L22.3821 55.5873L22.3921 55.457ZM40.6473 92.4498L45.0485 88.0485L43.0066 90.4079L40.806 92.6085L37.3463 95.7507L39.9384 92.8412L40.6473 92.4498ZM18.5042 48.7973L11.5457 55.7558L10.4249 56.3746L6.32684 60.9746L11.7967 56.0067L15.2759 52.5275L18.5042 48.7973ZM32.7113 78.139L31.1131 79.7372L30.8432 79.8668L29.9145 80.9358L31.1833 79.8074L31.9823 79.0083L32.7113 78.139ZM21.7577 93.9525L31.2855 84.0344L30.8324 84.8777L42.4999 73.2102L38.7408 77.2295L26.5552 89.6753L27.5914 88.1187L21.7577 93.9525ZM98.5132 90.0591L89.9224 97.9224L93.5769 94.9953L98.5132 90.0591ZM97.8456 80.2105L99.5027 78.6937L98.5506 79.6459L97.8456 80.2105ZM88.5656 56.4599L78.9205 65.7009L82.1262 63.3036L78.1413 67.2885L73.7522 70.8692L74.7195 70.5082L67.717 78.117L63.992 81.0336L58.0146 87.011L63.4289 81.7988L66.3887 79.4454L68.1212 78.5213L70.5757 75.6625L73.0302 72.8038L76.194 69.64L78.3434 67.4906L84.3208 61.5132L82.6575 62.7723L88.5656 56.4599ZM85.1893 67.0375L83.7304 68.356L84.3561 67.8707L85.1893 67.0375ZM90.7969 58.2022L99.2725 50.5418L94.4317 55.3826L90.7969 58.2022ZM79.377 76.2172L77.9182 77.5357L78.5438 77.0504L79.377 76.2172ZM59.4922 91.7253L56.4011 94.1231L60.0049 90.8659L63.6087 87.6087L59.4922 91.7253ZM63.8833 75.4153L46 92.3896L49.6884 89.1193L53.3767 85.8491L63.8833 75.4153ZM71.6063 55.0765L69.6609 57.0219L69.3475 57.1949L68.2018 58.481L69.731 57.0921L70.7037 56.1194L71.6063 55.0765ZM55.1405 71.6857L61.4131 65.4131L57.958 69.1267L55.1405 71.6857ZM65.8396 69.4497L61.7138 73.7138L64.2308 71.1968L63.7637 71.8484L69.0313 66.4886L70.6632 64.7645L65.6292 69.7985L65.8396 69.4497ZM53.0034 65.4955L58.2258 59.8914L58.0558 60.4431L64.5517 53.9472L62.5136 56.2398L55.7841 63.2238L56.2513 62.2475L53.0034 65.4955ZM97.0997 71.2032L79.6514 88.6515L86.7697 80.814L97.0997 71.2032ZM35.1848 56.2513L31.93 59.9006L34.0012 57.8294L33.804 58.5527L38.0451 54.0485L39.2945 52.5361L35.1519 56.6787L35.1848 56.2513ZM66.8712 26.2471L78.1907 14.3099L77.7244 15.394L91.6784 1.4399L87.233 6.29715L72.7096 21.2323L73.8482 19.2701L66.8712 26.2471ZM28.0473 68.2068L20.4355 76.375L25.1695 71.641L24.4884 73.0639L34.297 62.8844L37.2675 59.5429L27.7995 69.0109L28.0473 68.2068ZM8.94067 39.5658L14.1631 33.9617L13.993 34.5134L20.4889 28.0175L18.4509 30.3101L11.7213 37.2941L12.1886 36.3178L8.94067 39.5658ZM99.7403 26L88 37.7404L93.2735 32.9508L99.7403 26ZM1.93388 8.08743L4.77765 5.04974L4.67856 5.34275L8.20743 1.81388L7.09578 3.05481L3.4355 6.84437L3.69832 6.32299L1.93388 8.08743ZM54.4485 44.211L48.5985 50.061L47.6563 50.5813L44.211 54.4485L48.8095 50.272L51.7345 47.347L54.4485 44.211Z" />
	</pattern>`, luminanceCategory, fill)
}

func Rect(r *Runner, shape d2target.Shape) (string, error) {
	js := fmt.Sprintf(`node = rc.rectangle(0, 0, %d, %d, {
		fill: "#000",
		stroke: "#000",
		strokeWidth: %d,
		%s
	});`, shape.Width, shape.Height, shape.StrokeWidth, baseRoughProps)
	paths, err := computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	output := ""
	pathEl := svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill, pathEl.Stroke = svg_style.ShapeTheme(shape)
	pathEl.Class = "shape"
	pathEl.Style = shape.CSSStyle()
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	sketchOEl := svg_style.NewThemableElement("rect")
	sketchOEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	sketchOEl.Width = float64(shape.Width)
	sketchOEl.Height = float64(shape.Height)
	renderedSO, err := svg_style.NewThemableSketchOverlay(sketchOEl, pathEl.Fill).Render()
	if err != nil {
		return "", err
	}
	output += renderedSO

	return output, nil
}

func Oval(r *Runner, shape d2target.Shape) (string, error) {
	js := fmt.Sprintf(`node = rc.ellipse(%d, %d, %d, %d, {
		fill: "#000",
		stroke: "#000",
		strokeWidth: %d,
		%s
	});`, shape.Width/2, shape.Height/2, shape.Width, shape.Height, shape.StrokeWidth, baseRoughProps)
	paths, err := computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	output := ""
	pathEl := svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill, pathEl.Stroke = svg_style.ShapeTheme(shape)
	pathEl.Class = "shape"
	pathEl.Style = shape.CSSStyle()
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	soElement := svg_style.NewThemableElement("ellipse")
	soElement.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X+shape.Width/2, shape.Pos.Y+shape.Height/2)
	soElement.Rx = float64(shape.Width / 2)
	soElement.Ry = float64(shape.Height / 2)
	renderedSO, err := svg_style.NewThemableSketchOverlay(
		soElement,
		pathEl.Fill,
	).Render()
	if err != nil {
		return "", err
	}
	output += renderedSO

	return output, nil
}

// TODO need to personalize this per shape like we do in Terrastruct app
func Paths(r *Runner, shape d2target.Shape, paths []string) (string, error) {
	output := ""
	for _, path := range paths {
		js := fmt.Sprintf(`node = rc.path("%s", {
		fill: "#000",
		stroke: "#000",
		strokeWidth: %d,
		%s
	});`, path, shape.StrokeWidth, baseRoughProps)
		sketchPaths, err := computeRoughPathData(r, js)
		if err != nil {
			return "", err
		}
		pathEl := svg_style.NewThemableElement("path")
		pathEl.Fill, pathEl.Stroke = svg_style.ShapeTheme(shape)
		pathEl.Class = "shape"
		pathEl.Style = shape.CSSStyle()
		for _, p := range sketchPaths {
			pathEl.D = p
			output += pathEl.Render()
		}

		soElement := svg_style.NewThemableElement("path")
		for _, p := range sketchPaths {
			soElement.D = p
			renderedSO, err := svg_style.NewThemableSketchOverlay(
				soElement,
				pathEl.Fill,
			).Render()
			if err != nil {
				return "", err
			}
			output += renderedSO
		}
	}
	return output, nil
}

func Connection(r *Runner, connection d2target.Connection, path, attrs string) (string, error) {
	roughness := 1.0
	js := fmt.Sprintf(`node = rc.path("%s", {roughness: %f, seed: 1});`, path, roughness)
	paths, err := computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	output := ""
	animatedClass := ""
	if connection.Animated {
		animatedClass = " animated-connection"
	}

	pathEl := svg_style.NewThemableElement("path")
	pathEl.Fill = color.None
	pathEl.Stroke = svg_style.ConnectionTheme(connection)
	pathEl.Class = fmt.Sprintf("connection%s", animatedClass)
	pathEl.Style = connection.CSSStyle()
	pathEl.Attributes = attrs
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}
	return output, nil
}

// TODO cleanup
func Table(r *Runner, shape d2target.Shape) (string, error) {
	output := ""
	js := fmt.Sprintf(`node = rc.rectangle(0, 0, %d, %d, {
		fill: "#000",
		stroke: "#000",
		strokeWidth: %d,
		%s
	});`, shape.Width, shape.Height, shape.StrokeWidth, baseRoughProps)
	paths, err := computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	pathEl := svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill, pathEl.Stroke = svg_style.ShapeTheme(shape)
	pathEl.Class = "shape"
	pathEl.Style = shape.CSSStyle()
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	box := geo.NewBox(
		geo.NewPoint(float64(shape.Pos.X), float64(shape.Pos.Y)),
		float64(shape.Width),
		float64(shape.Height),
	)
	rowHeight := box.Height / float64(1+len(shape.SQLTable.Columns))
	headerBox := geo.NewBox(box.TopLeft, box.Width, rowHeight)

	js = fmt.Sprintf(`node = rc.rectangle(0, 0, %d, %f, {
		fill: "#000",
		%s
	});`, shape.Width, rowHeight, baseRoughProps)
	paths, err = computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	pathEl = svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill = shape.Fill
	pathEl.Class = "class_header"
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	if shape.Label != "" {
		tl := label.InsideMiddleLeft.GetPointOnBox(
			headerBox,
			20,
			float64(shape.LabelWidth),
			float64(shape.LabelHeight),
		)

		textEl := svg_style.NewThemableElement("text")
		textEl.X = tl.X
		textEl.Y = tl.Y + float64(shape.LabelHeight)*3/4
		textEl.Fill = shape.Stroke
		textEl.Class = "text"
		textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx",
			"start", 4+shape.FontSize,
		)
		textEl.Content = svg.EscapeText(shape.Label)
		output += textEl.Render()
	}

	var longestNameWidth int
	for _, f := range shape.Columns {
		longestNameWidth = go2.Max(longestNameWidth, f.Name.LabelWidth)
	}

	rowBox := geo.NewBox(box.TopLeft.Copy(), box.Width, rowHeight)
	rowBox.TopLeft.Y += headerBox.Height
	for _, f := range shape.Columns {
		nameTL := label.InsideMiddleLeft.GetPointOnBox(
			rowBox,
			d2target.NamePadding,
			rowBox.Width,
			float64(shape.FontSize),
		)
		constraintTR := label.InsideMiddleRight.GetPointOnBox(
			rowBox,
			d2target.TypePadding,
			0,
			float64(shape.FontSize),
		)

		textEl := svg_style.NewThemableElement("text")
		textEl.X = nameTL.X
		textEl.Y = nameTL.Y + float64(shape.FontSize)*3/4
		textEl.Fill = shape.PrimaryAccentColor
		textEl.Class = "text"
		textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx", "start", float64(shape.FontSize))
		textEl.Content = svg.EscapeText(f.Name.Label)
		output += textEl.Render()

		textEl.X = nameTL.X + float64(longestNameWidth) + 2*d2target.NamePadding
		textEl.Fill = shape.NeutralAccentColor
		textEl.Content = svg.EscapeText(f.Type.Label)
		output += textEl.Render()

		textEl.X = constraintTR.X
		textEl.Y = constraintTR.Y + float64(shape.FontSize)*3/4
		textEl.Fill = shape.SecondaryAccentColor
		textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx;letter-spacing:2px", "end", float64(shape.FontSize))
		textEl.Content = f.ConstraintAbbr()
		output += textEl.Render()

		rowBox.TopLeft.Y += rowHeight

		js = fmt.Sprintf(`node = rc.line(%f, %f, %f, %f, {
		%s
	});`, rowBox.TopLeft.X, rowBox.TopLeft.Y, rowBox.TopLeft.X+rowBox.Width, rowBox.TopLeft.Y, baseRoughProps)
		paths, err = computeRoughPathData(r, js)
		if err != nil {
			return "", err
		}
		pathEl := svg_style.NewThemableElement("path")
		pathEl.Fill = shape.Fill
		for _, p := range paths {
			pathEl.D = p
			output += pathEl.Render()
		}
	}

	sketchOEl := svg_style.NewThemableElement("rect")
	sketchOEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	sketchOEl.Width = float64(shape.Width)
	sketchOEl.Height = float64(shape.Height)
	renderedSO, err := svg_style.NewThemableSketchOverlay(sketchOEl, pathEl.Fill).Render()
	if err != nil {
		return "", err
	}
	output += renderedSO

	return output, nil
}

func Class(r *Runner, shape d2target.Shape) (string, error) {
	output := ""
	js := fmt.Sprintf(`node = rc.rectangle(0, 0, %d, %d, {
		fill: "#000",
		stroke: "#000",
		strokeWidth: %d,
		%s
	});`, shape.Width, shape.Height, shape.StrokeWidth, baseRoughProps)
	paths, err := computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	pathEl := svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill, pathEl.Stroke = svg_style.ShapeTheme(shape)
	pathEl.Class = "shape"
	pathEl.Style = shape.CSSStyle()
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	box := geo.NewBox(
		geo.NewPoint(float64(shape.Pos.X), float64(shape.Pos.Y)),
		float64(shape.Width),
		float64(shape.Height),
	)

	rowHeight := box.Height / float64(2+len(shape.Class.Fields)+len(shape.Class.Methods))
	headerBox := geo.NewBox(box.TopLeft, box.Width, 2*rowHeight)

	js = fmt.Sprintf(`node = rc.rectangle(0, 0, %d, %f, {
		fill: "#000",
		%s
	});`, shape.Width, headerBox.Height, baseRoughProps)
	paths, err = computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	pathEl = svg_style.NewThemableElement("path")
	pathEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	pathEl.Fill = shape.Fill
	pathEl.Class = "class_header"
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	sketchOEl := svg_style.NewThemableElement("rect")
	sketchOEl.Transform = fmt.Sprintf("translate(%d %d)", shape.Pos.X, shape.Pos.Y)
	sketchOEl.Width = float64(shape.Width)
	sketchOEl.Height = headerBox.Height
	renderedSO, err := svg_style.NewThemableSketchOverlay(sketchOEl, pathEl.Fill).Render()
	if err != nil {
		return "", err
	}
	output += renderedSO

	if shape.Label != "" {
		tl := label.InsideMiddleCenter.GetPointOnBox(
			headerBox,
			0,
			float64(shape.LabelWidth),
			float64(shape.LabelHeight),
		)

		textEl := svg_style.NewThemableElement("text")
		textEl.X = tl.X + float64(shape.LabelWidth)/2
		textEl.Y = tl.Y + float64(shape.LabelHeight)*3/4
		textEl.Fill = shape.Stroke
		textEl.Class = "text-mono"
		textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx",
			"middle",
			4+shape.FontSize,
		)
		textEl.Content = svg.EscapeText(shape.Label)
		output += textEl.Render()
	}

	rowBox := geo.NewBox(box.TopLeft.Copy(), box.Width, rowHeight)
	rowBox.TopLeft.Y += headerBox.Height
	for _, f := range shape.Fields {
		output += classRow(shape, rowBox, f.VisibilityToken(), f.Name, f.Type, float64(shape.FontSize))
		rowBox.TopLeft.Y += rowHeight
	}

	js = fmt.Sprintf(`node = rc.line(%f, %f, %f, %f, {
%s
	});`, rowBox.TopLeft.X, rowBox.TopLeft.Y, rowBox.TopLeft.X+rowBox.Width, rowBox.TopLeft.Y, baseRoughProps)
	paths, err = computeRoughPathData(r, js)
	if err != nil {
		return "", err
	}
	pathEl = svg_style.NewThemableElement("path")
	pathEl.Fill = shape.Fill
	pathEl.Class = "class_header"
	for _, p := range paths {
		pathEl.D = p
		output += pathEl.Render()
	}

	for _, m := range shape.Methods {
		output += classRow(shape, rowBox, m.VisibilityToken(), m.Name, m.Return, float64(shape.FontSize))
		rowBox.TopLeft.Y += rowHeight
	}

	return output, nil
}

func classRow(shape d2target.Shape, box *geo.Box, prefix, nameText, typeText string, fontSize float64) string {
	output := ""
	prefixTL := label.InsideMiddleLeft.GetPointOnBox(
		box,
		d2target.PrefixPadding,
		box.Width,
		fontSize,
	)
	typeTR := label.InsideMiddleRight.GetPointOnBox(
		box,
		d2target.TypePadding,
		0,
		fontSize,
	)

	textEl := svg_style.NewThemableElement("text")
	textEl.X = prefixTL.X
	textEl.Y = prefixTL.Y + fontSize*3/4
	textEl.Fill = shape.PrimaryAccentColor
	textEl.Class = "text-mono"
	textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx", "start", fontSize)
	textEl.Content = prefix
	output += textEl.Render()

	textEl.X = prefixTL.X + d2target.PrefixWidth
	textEl.Fill = shape.Fill
	textEl.Content = svg.EscapeText(nameText)
	output += textEl.Render()

	textEl.X = typeTR.X
	textEl.Y = typeTR.Y + fontSize*3/4
	textEl.Fill = shape.SecondaryAccentColor
	textEl.Style = fmt.Sprintf("text-anchor:%s;font-size:%vpx", "end", fontSize)
	textEl.Content = svg.EscapeText(typeText)
	output += textEl.Render()

	return output
}

func computeRoughPathData(r *Runner, js string) ([]string, error) {
	if _, err := r.run(js); err != nil {
		return nil, err
	}
	roughPaths, err := extractRoughPaths(r)
	if err != nil {
		return nil, err
	}
	return extractPathData(roughPaths)
}

func computeRoughPaths(r *Runner, js string) ([]roughPath, error) {
	if _, err := r.run(js); err != nil {
		return nil, err
	}
	return extractRoughPaths(r)
}

type attrs struct {
	D string `json:"d"`
}

type style struct {
	Stroke      string `json:"stroke,omitempty"`
	StrokeWidth string `json:"strokeWidth,omitempty"`
	Fill        string `json:"fill,omitempty"`
}

type roughPath struct {
	Attrs attrs `json:"attrs"`
	Style style `json:"style"`
}

func (rp roughPath) StyleCSS() string {
	style := ""
	if rp.Style.StrokeWidth != "" {
		style += fmt.Sprintf("stroke-width:%s;", rp.Style.StrokeWidth)
	}
	return style
}

func extractRoughPaths(r *Runner) ([]roughPath, error) {
	val, err := r.run("JSON.stringify(node.children, null, '  ')")
	if err != nil {
		return nil, err
	}

	var roughPaths []roughPath
	err = json.Unmarshal([]byte(val.String()), &roughPaths)
	if err != nil {
		return nil, err
	}

	// we want to have a fixed precision to the decimals in the path data
	for i := range roughPaths {
		// truncate all floats in path to only use up to 6 decimal places
		roughPaths[i].Attrs.D = floatRE.ReplaceAllStringFunc(roughPaths[i].Attrs.D, func(floatStr string) string {
			i := strings.Index(floatStr, ".")
			decimalLen := len(floatStr) - i - 1
			end := i + go2.Min(decimalLen, 6)
			return floatStr[:end+1]
		})
	}

	return roughPaths, nil
}

func extractPathData(roughPaths []roughPath) ([]string, error) {
	var paths []string
	for _, rp := range roughPaths {
		paths = append(paths, rp.Attrs.D)
	}
	return paths, nil
}

func ArrowheadJS(r *Runner, arrowhead d2target.Arrowhead, stroke string, strokeWidth int) (arrowJS, extraJS string) {
	// Note: selected each seed that looks the good for consistent renders
	switch arrowhead {
	case d2target.ArrowArrowhead:
		arrowJS = fmt.Sprintf(
			`node = rc.linearPath(%s, { strokeWidth: %d, stroke: "%s", seed: 3 })`,
			`[[-10, -4], [0, 0], [-10, 4]]`,
			strokeWidth,
			stroke,
		)
	case d2target.TriangleArrowhead:
		arrowJS = fmt.Sprintf(
			`node = rc.polygon(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "solid", seed: 2 })`,
			`[[-10, -4], [0, 0], [-10, 4]]`,
			strokeWidth,
			stroke,
			stroke,
		)
	case d2target.DiamondArrowhead:
		arrowJS = fmt.Sprintf(
			`node = rc.polygon(%s, { strokeWidth: %d, stroke: "%s", fill: "white", fillStyle: "solid", seed: 1 })`,
			`[[-20, 0], [-10, 5], [0, 0], [-10, -5], [-20, 0]]`,
			strokeWidth,
			stroke,
		)
	case d2target.FilledDiamondArrowhead:
		arrowJS = fmt.Sprintf(
			`node = rc.polygon(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "zigzag", fillWeight: 4, seed: 1 })`,
			`[[-20, 0], [-10, 5], [0, 0], [-10, -5], [-20, 0]]`,
			strokeWidth,
			stroke,
			stroke,
		)
	case d2target.CfManyRequired:
		arrowJS = fmt.Sprintf(
			// TODO why does fillStyle: "zigzag" error with path
			`node = rc.path(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "solid", fillWeight: 4, seed: 2 })`,
			`"M-15,-10 -15,10 M0,10 -15,0 M0,-10 -15,0"`,
			strokeWidth,
			stroke,
			stroke,
		)
	case d2target.CfMany:
		arrowJS = fmt.Sprintf(
			`node = rc.path(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "solid", fillWeight: 4, seed: 8 })`,
			`"M0,10 -15,0 M0,-10 -15,0"`,
			strokeWidth,
			stroke,
			stroke,
		)
		extraJS = fmt.Sprintf(
			`node = rc.circle(-20, 0, 8, { strokeWidth: %d, stroke: "%s", fill: "white", fillStyle: "solid", fillWeight: 1, seed: 4 })`,
			strokeWidth,
			stroke,
		)
	case d2target.CfOneRequired:
		arrowJS = fmt.Sprintf(
			`node = rc.path(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "solid", fillWeight: 4, seed: 2 })`,
			`"M-15,-10 -15,10 M-10,-10 -10,10"`,
			strokeWidth,
			stroke,
			stroke,
		)
	case d2target.CfOne:
		arrowJS = fmt.Sprintf(
			`node = rc.path(%s, { strokeWidth: %d, stroke: "%s", fill: "%s", fillStyle: "solid", fillWeight: 4, seed: 3 })`,
			`"M-10,-10 -10,10"`,
			strokeWidth,
			stroke,
			stroke,
		)
		extraJS = fmt.Sprintf(
			`node = rc.circle(-20, 0, 8, { strokeWidth: %d, stroke: "%s", fill: "white", fillStyle: "solid", fillWeight: 1, seed: 5 })`,
			strokeWidth,
			stroke,
		)
	}
	return
}

func Arrowheads(r *Runner, connection d2target.Connection, srcAdj, dstAdj *geo.Point) (string, error) {
	arrowPaths := []string{}

	if connection.SrcArrow != d2target.NoArrowhead {
		arrowJS, extraJS := ArrowheadJS(r, connection.SrcArrow, connection.Stroke, connection.StrokeWidth)
		if arrowJS == "" {
			return "", nil
		}

		startingSegment := geo.NewSegment(connection.Route[0], connection.Route[1])
		startingVector := startingSegment.ToVector().Reverse()
		angle := startingVector.Degrees()

		transform := fmt.Sprintf(`transform="translate(%f %f) rotate(%v)"`,
			startingSegment.Start.X+srcAdj.X, startingSegment.Start.Y+srcAdj.Y, angle,
		)

		roughPaths, err := computeRoughPaths(r, arrowJS)
		if err != nil {
			return "", err
		}
		if extraJS != "" {
			extraPaths, err := computeRoughPaths(r, extraJS)
			if err != nil {
				return "", err
			}
			roughPaths = append(roughPaths, extraPaths...)
		}

		pathEl := svg_style.NewThemableElement("path")
		pathEl.Class = "connection"
		pathEl.Attributes = transform
		for _, rp := range roughPaths {
			pathEl.D = rp.Attrs.D
			pathEl.Fill = rp.Style.Fill
			pathEl.Stroke = rp.Style.Stroke
			pathEl.Style = rp.StyleCSS()
			arrowPaths = append(arrowPaths, pathEl.Render())
		}
	}

	if connection.DstArrow != d2target.NoArrowhead {
		arrowJS, extraJS := ArrowheadJS(r, connection.DstArrow, connection.Stroke, connection.StrokeWidth)
		if arrowJS == "" {
			return "", nil
		}

		length := len(connection.Route)
		endingSegment := geo.NewSegment(connection.Route[length-2], connection.Route[length-1])
		endingVector := endingSegment.ToVector()
		angle := endingVector.Degrees()

		transform := fmt.Sprintf(`transform="translate(%f %f) rotate(%v)"`,
			endingSegment.End.X+dstAdj.X, endingSegment.End.Y+dstAdj.Y, angle,
		)

		roughPaths, err := computeRoughPaths(r, arrowJS)
		if err != nil {
			return "", err
		}
		if extraJS != "" {
			extraPaths, err := computeRoughPaths(r, extraJS)
			if err != nil {
				return "", err
			}
			roughPaths = append(roughPaths, extraPaths...)
		}

		pathEl := svg_style.NewThemableElement("path")
		pathEl.Class = "connection"
		pathEl.Attributes = transform
		for _, rp := range roughPaths {
			pathEl.D = rp.Attrs.D
			pathEl.Fill = rp.Style.Fill
			pathEl.Stroke = rp.Style.Stroke
			pathEl.Style = rp.StyleCSS()
			arrowPaths = append(arrowPaths, pathEl.Render())
		}
	}

	return strings.Join(arrowPaths, " "), nil
}
