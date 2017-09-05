#include <windows.h>
#include <commctrl.h>
#include <stdio.h>

#include "Aggregator.h"
#include "Error.h"
#include "Slotter.h"
#include "CounterManager.h"


#define SRV_MOUSE_DLGID 102
#define SRV_PROGRESS_BARID 1000
#define SRV_MOUSE_COLLECTID 1002
#define SRV_MOUSE_STATIC 1003
#define SRV_TEXT_A 1004
#define SRV_TEXT_B 1005
#define SRV_TEXT_C 1006

DWORD WINAPI AcceptThread(LPVOID lpParameter);

static BOOL CALLBACK KeyboardEntropyProc(HWND hwndDlg, UINT uMsg, WPARAM wParam, LPARAM)
{
    BOOL ret = FALSE;

    if ((uMsg == WM_COMMAND) && (HIWORD(wParam) == BN_CLICKED)) {
        if (LOWORD(wParam) == IDCANCEL) {
            BOOL* pQuit = (BOOL*)GetWindowLong(hwndDlg, DWL_USER);
            DestroyWindow(hwndDlg);
            *pQuit = TRUE;
            ret = TRUE;
        }
    }

    return ret;
}


HWND DialogInit(HINSTANCE hInstance, HWND hWndParent, BOOL* quit)
{
    HWND ret = 0;
    ret = CreateDialog(hInstance, MAKEINTRESOURCE(SRV_MOUSE_DLGID), hWndParent, KeyboardEntropyProc);

    if (!ret) {
        goto error;
    }

    if (hWndParent) {
        EnableWindow(hWndParent, FALSE);
    }

    SetWindowLong(ret, DWL_USER, (LONG)quit);
error:
    return ret;
}

void ProgressInit(HWND hwndDlg, int param)
{
    SendDlgItemMessage(hwndDlg, SRV_PROGRESS_BARID, PBM_SETRANGE32, 0, param);
}

void ProgressSet(HWND hwndDlg, double param)
{
    SendDlgItemMessage(hwndDlg, SRV_PROGRESS_BARID,
                       PBM_SETPOS, (WPARAM)param, 0);
}


void TextSetStr(HWND hwndDlg, DWORD id, TCHAR* txt)
{
    int const arraysize = 160;
    TCHAR pszTmp[arraysize];
    GetDlgItemText(hwndDlg, id, pszTmp, arraysize);

    if (strcmp(pszTmp, txt) != 0) {
        SetDlgItemText(hwndDlg, id, txt);
    }
}

void TextSet(HWND hwndDlg, DWORD avail, Counter cntr)
{
    int const arraysize = 160;
    TCHAR pszDestA[arraysize];
    TCHAR pszDestB[arraysize];
    TCHAR pszDestC[arraysize];
    size_t cbDest = arraysize * sizeof(TCHAR);
    LPCTSTR pszFormatA = TEXT("Available: %d");
    LPCTSTR pszFormatB = TEXT("All, Current: %d, Max: %d");
    LPCTSTR pszFormatC = TEXT("Good, Current: %d, Max: %d");
    snprintf(pszDestA, cbDest, pszFormatA, avail);
    snprintf(pszDestB, cbDest, pszFormatB, cntr.lastRequest, cntr.maxRequest);
    snprintf(pszDestC, cbDest, pszFormatC, cntr.lastSuccessRequest,
             cntr.maxSuccessRequest);
    TextSetStr(hwndDlg, SRV_TEXT_A, pszDestA);
    TextSetStr(hwndDlg, SRV_TEXT_B, pszDestB);
    TextSetStr(hwndDlg, SRV_TEXT_C, pszDestC);
}


BOOL HandleMouseMove(HWND hwndDlg, MSG msg, Aggregator* dd)
{
    MousePosition mpos;
    mpos.ptMousePos.x = LOWORD(msg.lParam);
    mpos.ptMousePos.y = HIWORD(msg.lParam);
    mpos.dwTickCount = GetTickCount();
    ClientToScreen(hwndDlg, &(mpos.ptMousePos));
    return dd->handle(&mpos);
}

BOOL HandleKeyPress(MSG msg, Aggregator* dd)
{
    KeyPressed key;
    key.bScanCode = ((msg.lParam >> 16) & 0x0000000F);
    key.dwTickCount = GetTickCount();
    key.isUp = (msg.message == WM_KEYUP);
    return dd->handle(&key);
}

BOOL GatherEntropy(HINSTANCE hInstance, HWND hWndParent)
{
    MSG msg;
    BOOL quit = FALSE;
    BOOL bResult = FALSE;
    Aggregator* aggr = NULL;
    aggr = new Aggregator();
    HWND hwndDlg = DialogInit(hInstance, hWndParent, &quit);

    if (!hwndDlg) {
        goto error;
    }

    if (!aggr->init()) {
        goto error;
    }

    ProgressInit(hwndDlg, aggr->requested());
    TextSet(hwndDlg, 0, Counter());

    while ((quit == FALSE) && GetMessage(&msg, 0, 0, 0) > 0) {
        if (!aggr->enoughEntropy()) {
            switch (msg.message) {
            case WM_MOUSEMOVE:
                if (HandleMouseMove(hwndDlg, msg, aggr)) {
                    ProgressSet(hwndDlg, aggr->entropy());
                }

                break;

            case WM_KEYDOWN:
            case WM_KEYUP:
                if (HandleKeyPress(msg, aggr)) {
                    ProgressSet(hwndDlg, aggr->entropy());
                }
            }
        }

        if (!IsDialogMessage(hwndDlg, &msg)) {
            TranslateMessage(&msg);
            DispatchMessage(&msg);
        }

        if (aggr->enoughEntropy()) {
            aggr->prepareSlice();
            Slotter::push(aggr->getSlice());
            delete aggr;
            aggr = NULL;
            aggr = new Aggregator();
            aggr->init();
            ProgressInit(hwndDlg, aggr->requested());
        }

        DWORD avail = Slotter::available();
        Counter cntr = CounterManager::get();
        TextSet(hwndDlg, avail, cntr);
    }

    if (hWndParent) {
        EnableWindow(hWndParent, TRUE);
    }

    bResult = TRUE;
error:

    if (aggr) {
        delete aggr;
        aggr = NULL;
    }

    return bResult;
}


int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE, LPSTR, int)
{
    INITCOMMONCONTROLSEX commonControls;
    HANDLE threadHandle = NULL;
    DWORD threadId = 0;
    commonControls.dwSize = sizeof(commonControls);
    commonControls.dwICC = ICC_PROGRESS_CLASS;
    InitCommonControlsEx(&commonControls);

    if (!Aggregator::initialize()) {
        OutputError("Aggregator::initialize", GetLastError());
        goto error;
    }

    CounterManager::initialize();
    Slotter::initialize(Aggregator::outbytes());
    threadHandle = CreateThread(NULL, 0, AcceptThread, 0, 0, &threadId);

    if (threadHandle == NULL) {
        OutputError("CreateThread()", GetLastError());
        goto error;
    }

    if (!GatherEntropy(hInstance, 0)) {
        OutputError("GatherEntropy()", GetLastError());
    }

error:
    Aggregator::finalize();
    Slotter::finalize();
    return 0;
}

