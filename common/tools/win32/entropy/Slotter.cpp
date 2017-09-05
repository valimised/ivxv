#include "Slotter.h"

#include <assert.h>
#include <cstddef>
#include <windows.h>


CRITICAL_SECTION CriticalSectionSlotter;

Slotter* Slotter::impl = NULL;

void Slotter::initialize(int slen)
{
    InitializeCriticalSection(&CriticalSectionSlotter);
    impl = new Slotter(slen);
}

void Slotter::finalize()
{
    DeleteCriticalSection(&CriticalSectionSlotter);
    delete impl;
    impl = NULL;
}

void Slotter::acquire()
{
    EnterCriticalSection(&CriticalSectionSlotter);
}

void Slotter::release()
{
    LeaveCriticalSection(&CriticalSectionSlotter);
}

void Slotter::push(BYTE* data)
{
    acquire();
    impl->doPush(data);
    release();
}

DWORD Slotter::available()
{
    DWORD ret = 0;
    acquire();
    ret = impl->dwAvailable;
    release();
    return ret;
}

bool Slotter::request(BYTE* buffer, DWORD count)
{
    bool ret = false;
    acquire();
    ret = impl->doRequest(buffer, count);
    release();
    return ret;
}

Slotter::Slotter(int slen) : slots()
{
    dwAvailable = 0;
    slice = NULL;
    slice_avail = 0;
    slice_length = slen;
}

Slotter::~Slotter()
{
    if (slice != NULL) {
        LocalFree(slice);
        slice = NULL;
    }

    while (!slots.empty()) {
        slice = slots.front();
        slots.pop_front();
        LocalFree(slice);
        slice = NULL;
    }
}

void Slotter::doPush(BYTE* data)
{
    if (data) {
        dwAvailable += slice_length;
        slots.push_back(data);
    }
}

void Slotter::updateSlice()
{
    if (slice_avail == 0) {
        assert(slice == NULL);

        if (!slots.empty()) {
            slice = slots.front();
            slots.pop_front();
            slice_avail = slice_length;
        }
    }
}

int Slotter::copyFromSlice(BYTE* dst, int required)
{
    int tobecopied = slice_avail;

    if (required < slice_avail) {
        tobecopied = required;
    }

    if (slice_avail > 0) {
        assert(slice != NULL);
        memcpy(dst, slice + (slice_length - slice_avail), tobecopied);
        slice_avail -= tobecopied;
    }

    if (slice_avail == 0) {
        if (slice != NULL) {
            LocalFree(slice);
            slice = NULL;
        }
    }

    return tobecopied;
}

bool Slotter::doRequest(BYTE* buffer, DWORD count)
{
    if (count > dwAvailable) {
        return false;
    }

    if (buffer) {
        BYTE* tmp = buffer;
        DWORD gathered = 0;

        while (gathered < count) {
            updateSlice();
            int inc = copyFromSlice(tmp, count - gathered);
            gathered += inc;
            tmp += inc;
        }

        dwAvailable -= count;
    }

    return true;
}

