package dandler

import (
	"net/http"

	"github.com/jakdept/drings"
)

// Header returns a handler that adds the given handler to the response.
func Header(name, msg string, handler http.Handler) http.Handler {
	return headerHandler{key: name, value: msg, child: handler}
}

type headerHandler struct {
	key   string
	value string
	child http.Handler
}

func (h headerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(h.key, h.value)
	h.child.ServeHTTP(w, r)
}

// ASCIIHeader adds a multiline header to the response - a header for each line.
// It is used to add ascii art to the header of a response.
//
// Both key and value can be a multiline string = and will be broken up as
// needed. Key will be repeated for the duration of value for each header line.
//
// You need to make sure both key and value have a consistent length after white space
// trimming on each line, as HTTP will do that anyway. If you don't, this will
// split on newlines, trim spaces, then right pad the key with underscores so
// things line up anyway.
func ASCIIHeader(key, value, padChar string, child http.Handler) http.Handler {
	keyArr := drings.SplitAndTrimSpace(key, "\n")
	width := drings.MaxLen(keyArr)
	keyArr = drings.PadAllRight(keyArr, padChar, width)

	valueArr := drings.SplitAndTrimSpace(value, "\n")

	return asciiHeaderHandler{key: keyArr, value: valueArr, child: child}
}

type asciiHeaderHandler struct {
	key   []string
	value []string
	child http.Handler
}

func (h asciiHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keyCount, valueCount := 0, 0
	for valueCount < len(h.value) {
		if keyCount >= len(h.key) {
			keyCount = 0
		}
		w.Header().Add(h.key[keyCount], h.value[valueCount])
		keyCount++
		valueCount++
	}
	h.child.ServeHTTP(w, r)
}

// GolangGopherASCII is a ascii drawing of a golang gopher for use in ASCIIHeader()
const GolangGopherASCII = `
'______________________________________________________________________________
/                                                                              \
|                                                                              |
|                                 '-/+++oo++/-'      ''                        |
|                            '/ooso+/:/ossosyysso/+ssooss+                     |
|                         :oss/:::::oy+.     -oy//syoos/:+h'                   |
|                      :sy+::::::::y+       '/o/yo::+hMM/:s+                   |
|                   -sy+::::::::::oy        dMMsyd/:::/hs:h-                   |
|                 :yo::+osoo/:::::s+        :hmh-y+:::::+do                    |
|               .yo:os+-.''./oy/::/d'           .d::::::::yo                   |
|              -h//h:       :++sy::/h:         +h::::::::::oy'                 |
|             .d::d.       oMMhhoo:::+so+/:/+os+::::::::::::+h'                |
|           '-yo::m        -hmd/-h:sdmNmyss+:::::::::::::::::+y                |
|        'oyo+N:::d-            ssNMMMMMh+oyd:::::::::::::::::y+               |
|        h+:/yN:::/h:         'syydddhs++++sd::::::::::::::::::d'              |
|       'm::mMM/::::oso:.''-/oy/ys+++oyds+om:::::::::::::::::::oy'             |
|        yo:/sds:::::::/ooo+/:::odsyhy'.h' :h:::::::::::::::::::/h-            |
|         +ys++m::::::::::::::::::/::d' -d/+h:::::::::::::::::::::d-           |
|           '-:ss::::::::::::::::::::/yoso/::::::::::::::::::::::::do++++'     |
|              'd/:::::::::::::::::::::::::::::::::::::::::::::::::oy -+ho     |
|               -d::::::::::::::::::::::::::::::::::::::::::::::::::do++o'     |
|                /h:::::::::::::::::::::::::::::::::::::::::::::::::os         |
|                 /h::::::::::::::::::::::::::::::s::::::::::::::::::m         |
|                  :h::::::::::::::::::::::::::::+mosso/:::::::::::::d.        |
|                   -h/:::::::::::::::::::::::::/d-  ++m:::::::::::::y/        |
|                    'yo::::::::::::::::::::::+yy+sso+yy:::::::::::::os        |
|                      oy:::::::::::::::::::::/::::::::::::::::::::::+y        |
|                       os:::::::::::::::::::::::::::::::::::::::::::/h        |
|                        h+::::::::::::::::::::::::::::::::::::::::::/h        |
|                        .d::::::::::::::::::::::::::::::::::::::::::/h        |
|                         s+:::::::::::::::::::::::::::::::::::::::::+y        |
|                         :h:::::::::::::::::::::::::::::::::::::::::oo        |
|                         'm:::::::::::::::::::::::::::::::::::::::::y:        |
|                          m:::::::::::::::::::::::::::::::::::::::::m'        |
|                          m::::::::::::::::::::::::::::::::::::::::+y         |
|                          m::::::::::::::::::::::::::::::::::::::::m/'        |
|                          m:::::::::::::::::::::::::::::::::::::::h/-/oo'     |
|                          y+::::::::::::::::::::::::::::::::::::/do' .oom     |
|                          :h::::::::::::::::::::::::::::::::::/ys.-+o+/h+     |
|                         'sd+::::::::::::::::::::::::::::::/ss+'      '       |
|                         +s+N+::::::::::::::::::::::::::/ss+.                 |
|                          '' oy/::::::::::::::::::::/oso/'                    |
|                              '/sso/::::::::::/+ssoo/.                        |
|                                -y-:++dsoooo++/-'                             |
|                                d.s  o+                                       |
|                                d+y'/y                                        |
|                                '/++-                                         |
|                                                                              |
\______________________________________________________________________________/
`
