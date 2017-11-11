package xsd

/*

#include <stdio.h>
#include <stdarg.h>

// The gateway function
void xmlErrorFunc_cgo(void *ctx, const char * msg, ...)
{
    int *id = (int*)ctx;
    void xmlErrorFunc(int, const char *);

    char buf[1024];
    va_list args;
    va_start(args, msg);
    int len = vsnprintf(buf, sizeof(buf), msg, args);
    va_end(args);

    xmlErrorFunc(*id, buf);
}
*/
import "C"
