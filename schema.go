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
	"unsafe"
)

type Schema struct {
	Ptr C.xmlSchemaPtr
}

type DocPtr C.xmlDocPtr

//export xmlErrorFunc
func xmlErrorFunc(id unsafe.Pointer, msg *C.char) {
	errs := (*SchemaErrors)(id)
	*errs = append(*errs, C.GoString(msg))
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

type SchemaErrors []string

func (e SchemaErrors) Error() string {
	return strings.Join(e, "")
}

// Validate uses its Schema to check an xml doc.  If the doc fails to match
// the schema, a list of errors is returned, nil otherwise.
func (s *Schema) Validate(doc DocPtr) SchemaErrors {
	validCtxt := C.xmlSchemaNewValidCtxt(s.Ptr)
	if validCtxt == nil {
		// TODO find error
		return SchemaErrors{"Could not build validator"}
	}
	defer C.xmlSchemaFreeValidCtxt(validCtxt)

	validationErrors := SchemaErrors{}

	C.xmlSchemaSetValidErrors(validCtxt,
		(C.xmlSchemaValidityErrorFunc)(unsafe.Pointer(C.xmlErrorFunc_cgo)),
		(C.xmlSchemaValidityErrorFunc)(unsafe.Pointer(C.xmlErrorFunc_cgo)),
		unsafe.Pointer(&validationErrors),
	)

	if C.xmlSchemaValidateDoc(validCtxt, doc) != 0 {
		return validationErrors
	}
	return nil
}
