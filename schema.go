package xsd

/*
#cgo pkg-config: libxml-2.0
#include <stdlib.h>
#include <libxml/xmlschemas.h>

extern void SchemaValidationErrorHandler(char *msg, xmlSchemaValidCtxtPtr *ctx, char *id);


static inline void HandleValidationError(void *ctx, const char *format, ...) {
	char *id;
	id = ctx;

    char *errMsg;
    va_list args;
    va_start(args, format);
    vasprintf(&errMsg, format, args);
    va_end(args);
	SchemaValidationErrorHandler(errMsg, ctx, id);
    free(errMsg);
}

static inline int XsdValidateSchema(xmlSchemaValidCtxtPtr ctxt,	xmlDocPtr doc, char *id){

	xmlSetStructuredErrorFunc(NULL, NULL);
    xmlSchemaSetValidErrors(ctxt,
                            HandleValidationError,
                            NULL,
                            id);
	return xmlSchemaValidateDoc(ctxt, doc);
}


*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type ErrorHandler func(string)

var (
	callbackMap map[string]ErrorHandler
)

type Schema struct {
	Ptr C.xmlSchemaPtr
}

type DocPtr C.xmlDocPtr

//export SchemaValidationErrorHandler
func SchemaValidationErrorHandler(msg *C.char, ctx *C.xmlSchemaValidCtxtPtr, id *C.char) {
	// input is an error message as pointer to char array
	errorMessage := C.GoString(msg)
	cid := C.GoString(id)

	if handler, ok := callbackMap[cid]; ok == true {
		handler(errorMessage)
	}
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
		return nil, errors.New("Could not parse schema")
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
func (s *Schema) Validate(doc DocPtr, handler ErrorHandler) error {
	if callbackMap == nil {
		callbackMap = make(map[string]ErrorHandler)
	}

	validCtxt := C.xmlSchemaNewValidCtxt(s.Ptr)
	if validCtxt == nil {
		// TODO find error - see below
		return errors.New("Could not build validator")
	}
	defer C.xmlSchemaFreeValidCtxt(validCtxt)

	contextAddr := fmt.Sprintf("%p", &validCtxt)
	cContextAddr := C.CString(contextAddr)

	callbackMap[contextAddr] = handler
	defer delete(callbackMap, contextAddr)

	result := int(C.XsdValidateSchema(validCtxt, doc, cContextAddr))
	if result != 0 {
		return errors.New("Document validation error")
	}

	return nil
}
