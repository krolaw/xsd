package xsd

/*
#cgo pkg-config: libxml-2.0
//#include <libxml/valid.h>
#include <libxml/xmlschemas.h>
//#include <stdio.h>

//static inline xmlSchemaValidityErrorFunc *getSchemaValidityErrorFunc() { return schemaValidityErrorFunc; }
//static inline void free_string(char* s) { free(s); }
//static inline xmlChar *to_xmlcharptr(const char *s) { return (xmlChar *)s; }
//static inline char *to_charptr(const xmlChar *s) { return (char *)s; }
//void xmlSchemaValidityErrorFunc	(void * ctx, const char * ms, ... )

*/
import "C"

import (
	//"github.com/jbussdieker/golibxml"

	"errors"
	//"fmt"
	"runtime"
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
		return errors.New("Document validation error")
	}
	return nil
}
