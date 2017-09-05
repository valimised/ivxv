// Command Line Options:
//    client [-p:x] [-s:IP] [-n:x] [-o]
//           -p:x      Remote port to send to
//           -s:IP     Server's IP address or hostname
//           -n:x      Number of bytes to enquire
#include <ctype.h>
#include <winsock2.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#define DEFAULT_COUNT       20
#define DEFAULT_PORT        5150
#define DEFAULT_BUFFER      2048

char  szServer[128];               // Server to connect to
int   iPort     = DEFAULT_PORT;    // Port on server to connect to
DWORD dwCount   = DEFAULT_COUNT;   // Number of bytes to enquire


void usage()
{
    printf("client [-p:x] [-s:IP] [-n:x]\n\n");
    printf("       -p:x      Remote port to send to\n");
    printf("       -s:IP     Server's IP address or hostname\n");
    printf("       -n:x      Number of bytes to enquire\n");
    printf("\n");
}


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


void ValidateArgs(int argc, char** argv)
{
    for (int i = 1; i < argc; i++) {
        if ((argv[i][0] == '-') || (argv[i][0] == '/')) {
            switch (tolower(argv[i][1])) {
            case 'p':        // Remote port
                if (strlen(argv[i]) > 3) {
                    iPort = atoi(&argv[i][3]);
                }

                break;

            case 's':       // Server
                if (strlen(argv[i]) > 3) {
                    strcpy(szServer, &argv[i][3]);
                }

                break;

            case 'n':       // Number of times to send message
                if (strlen(argv[i]) > 3) {
                    dwCount = (DWORD)atol(&argv[i][3]);
                }

                break;

            default:
                usage();
                break;
            }
        }
    }
}

int main(int argc, char** argv)
{
    WSADATA wsd;
    SOCKET sClient;
    char szBuffer[DEFAULT_BUFFER];
    int ret;
    struct sockaddr_in server;
    struct hostent* host = NULL;

    if (argc < 2) {
        usage();
        exit(1);
    }

    // Parse the command line and load Winsock
    ValidateArgs(argc, argv);

    if (WSAStartup(MAKEWORD(2, 2), &wsd) != 0) {
        OutputError("WSAStartup()", WSAGetLastError());
        return 1;
    }

    // Create the socket, and attempt to connect to the server
    sClient = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);

    if (sClient == INVALID_SOCKET) {
        OutputError("socket()", WSAGetLastError());
        return 1;
    }

    server.sin_family = AF_INET;
    server.sin_port = htons((unsigned short)iPort);
    server.sin_addr.s_addr = inet_addr(szServer);

    if (server.sin_addr.s_addr == INADDR_NONE) {
        host = gethostbyname(szServer);

        if (host == NULL) {
            printf("Unable to resolve server %s\n", szServer);
            return 1;
        }

        CopyMemory(&server.sin_addr, host->h_addr_list[0], (size_t)host->h_length);
    }

    ret = connect(sClient, (struct sockaddr*)&server, sizeof(server));

    if (ret == SOCKET_ERROR) {
        OutputError("connect()", WSAGetLastError());
        return 1;
    }

    DWORD out = htonl(dwCount);
    ret = send(sClient, (const char*)&out, sizeof(out), 0);

    if (ret == 0) {
        return 1;
    }
    else if (ret == SOCKET_ERROR) {
        OutputError("send()", WSAGetLastError());
        return 1;
    }

    ret = recv(sClient, szBuffer, DEFAULT_BUFFER, 0);

    if (ret == 0) {
        printf("It is a graceful close!\n");
    }
    else if (ret == SOCKET_ERROR) {
        OutputError("recv()", WSAGetLastError());
    }
    else {
        printf("Received %d bytes:\n", ret);
        char resptype = szBuffer[0];

        if (ret == 1) {
            printf("Entroy provider would block\n");
            assert(resptype == 0);
        }
        else {
            assert(resptype == (char)0xFF);
            assert((unsigned int)ret == (dwCount + 1));
        }

        for (int i = 1; i < ret; i++) {
            printf("%x:", (unsigned char)szBuffer[i]);
        }

        printf("\n");
    }

    if (closesocket(sClient) != 0) {
        OutputError("closesocket()", WSAGetLastError());
    }

    if (WSACleanup() != 0) {
        OutputError("WSACleanup()", WSAGetLastError());
    }

    return 0;
}
