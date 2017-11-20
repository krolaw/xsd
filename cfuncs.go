package xsd

/*
#include <stdio.h>
#include <stdarg.h>

// The gateway function
void xmlErrorFunc_cgo(void *ctx, const char * format, ...)
{
    void xmlErrorFunc(void *ctx, const char *);

    char buf[1024];
    va_list args;
    va_start(args, format);
    vsnprintf(buf, sizeof(buf), format, args);
    va_end(args);

    xmlErrorFunc(ctx, buf);
}
*/
import "C"
