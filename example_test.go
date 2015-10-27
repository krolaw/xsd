// Example of schema check
package xsd

import (
	// The xsd package does not read in documents.
	// Any package that exposes the document's libxml2's xmlDocPtr will do.
	// This one seemed rather straight forward, but others should be fine.
	"github.com/jbussdieker/golibxml"
	"github.com/krolaw/xsd"

	"fmt"
	"unsafe"
)

const (
	XSD = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<xs:element name="DayCount">
		<xs:simpleType>
		<xs:restriction base="xs:int">
			<xs:minInclusive value="0" />
			<xs:maxInclusive value="9999" />
		</xs:restriction>
		</xs:simpleType>
	</xs:element>
</xs:schema>`

	XML = `<DayCount>-1</DayCount>` // This XML is invalid (value < 0)
)

func ExampleSchema_Validate() {

	xsdSchema, err := xsd.ParseSchema([]byte(XSD))
	if err != nil {
		fmt.Println(err)
		return
	}

	doc := golibxml.ParseDoc(XML)
	if doc == nil {
		// TODO capture and display error - help please
		fmt.Println("Error parsing document")
		return
	}
	defer doc.Free()

	// golibxml._Ctype_xmlDocPtr can't be cast to xsd.DocPtr, even though they are both
	// essentially _Ctype_xmlDocPtr.  Using unsafe gets around this.
	if err := xsdSchema.Validate(xsd.DocPtr(unsafe.Pointer(doc.Ptr)), func(x string) { fmt.Printf("Message from handler #1: %s\n", x) }); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("XML Valid as per XSD")
}
