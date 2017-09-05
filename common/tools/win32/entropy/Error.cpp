#include <windows.h>
#include <stdio.h>

void OutputError(const char* prefix, int err)
{
    char* s = NULL;
    FormatMessage(FORMAT_MESSAGE_ALLOCATE_BUFFER | FORMAT_MESSAGE_FROM_SYSTEM |
                  FORMAT_MESSAGE_IGNORE_INSERTS, NULL, (DWORD)err,
                  MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
                  (LPSTR)&s, 0, NULL);
    fprintf(stderr, "%s\ncode: %d\n%s\n", prefix, err, s);
    LocalFree(s);
}

