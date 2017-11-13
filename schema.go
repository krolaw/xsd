package xsd

/*
#cgo pkg-config: libxml-2.0
#include <libxml/xmlschemas.h>

void xmlErrorFunc_cgo(void *, const char *); // Forward declaration.
*/
import "C"

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

type Schema struct {
	Ptr C.xmlSchemaPtr
}

type DocPtr C.xmlDocPtr

/*
//export schemaValidityErrorFunc
// This was one idea I wanted to use to extract the error.  It doesn't work
// because ...[]interface{} makes the func unexportable.  Not sure what to
// replace this with. - help please
func schemaValidityErrorFunc(ctx unsafe.Pointer, format *C.char, values ...[]interface{}) {
	*(*error)(ctx) = fmt.Errorf(C.GoString(format), values...)
}*/

var validationErrorsMu sync.Mutex
var validationErrors = map[int][]string{}
var validationErrorsNextIndex = 0

//export xmlErrorFunc
func xmlErrorFunc(id int, msg *C.char) {
	validationErrorsMu.Lock()
	validationErrors[id] = append(validationErrors[id], C.GoString(msg))
	validationErrorsMu.Unlock()
}

// ParseSchema creates new Schema from []byte containing xml schema data.
// Will probably change []byte to DocPtr.
func ParseSchema(buffer []byte) (*Schema, error) {
	cSchemaNewMemParserCtxt := C.xmlSchemaNewMemParserCtxt((*C.char)(unsafe.Pointer(&buffer[0])), C.int(len(buffer)))
	if cSchemaNewMemParserCtxt == nil {
		return nil, errors.New("Could not create schema parser") // TODO extract error - see Validate func below
	}
	defer C.xmlSchemaFreeParserCtxt(cSchemaNewMemParserCtxt)
	cSchema := C.xmlSchemaParse(cSchemaNewMemParserCtxt)
	if cSchema == nil {
		return nil, errors.New("Could not parse schema") // TODO extract error - see Validate func below
	}
	return makeSchema(cSchema), nil
}

func finaliseSchema(s *Schema) {
	C.xmlSchemaFree(s.Ptr)
}

func makeSchema(cSchema C.xmlSchemaPtr) *Schema {
	s := &Schema{cSchema}
	runtime.SetFinalizer(s, finaliseSchema)
	return s
}

// Validate uses its Schema to check an xml doc.  If the doc fails to match
// the schema, an error is returned, nil otherwise.
// At the moment, the error just says that the document failed.  It doesn't
// say where and why.  It needs to - help greatly appreciated.
func (s *Schema) Validate(doc DocPtr) error {
	validCtxt := C.xmlSchemaNewValidCtxt(s.Ptr)
	if validCtxt == nil {
		// TODO find error - see below
		return errors.New("Could not build validator")
	}
	defer C.xmlSchemaFreeValidCtxt(validCtxt)

	validationErrorsMu.Lock()
	validationErrorsNextIndex++
	id := validationErrorsNextIndex
	validationErrors[id] = []string{}
	validationErrorsMu.Unlock()
	defer func() {
		validationErrorsMu.Lock()
		delete(validationErrors, id)
		validationErrorsMu.Unlock()
	}()
	C.xmlSchemaSetValidErrors(validCtxt,
		(C.xmlSchemaValidityErrorFunc)(unsafe.Pointer(C.xmlErrorFunc_cgo)),
		(C.xmlSchemaValidityErrorFunc)(unsafe.Pointer(C.xmlErrorFunc_cgo)),
		unsafe.Pointer(&id),
	)
	/*
		// My plan was to register my go func to receive errors and pass it an
		// error ptr specific to this validation (useful for multiple goroutines).
		// Alas it doesn't work, help appreciated.

		var err *error
		if C.xmlSchemaGetValidErrors(validCtxt, C.schemaValidityErrorFunc, nil, unsafe.Pointer(err)) == -1 {
			return errors.New("Could not set error func.")
		}
	*/

	if C.xmlSchemaValidateDoc(validCtxt, doc) != 0 {
		//return errors.New(*err) // When the above works
		return errors.New(strings.Join(validationErrors[id], ""))
	}
	return nil
}
