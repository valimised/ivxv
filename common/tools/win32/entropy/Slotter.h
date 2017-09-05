#pragma once

#include <WinDef.h>
#include <list>

class Slotter {

public:

    static void initialize(int);
    static void finalize();
    static void push(BYTE*);
    static DWORD available();
    static bool request(BYTE* buffer, DWORD count);

protected:

    static void acquire();
    static void release();

    Slotter(int);
    virtual ~Slotter();

    void doPush(BYTE*);
    bool doRequest(BYTE* buffer, DWORD count);
    void updateSlice();
    int copyFromSlice(BYTE* dst, int required);

    DWORD dwAvailable;
    std::list<BYTE*> slots;

    BYTE* slice;
    int slice_avail;
    int slice_length;

private:

    static Slotter* impl;
    Slotter(const Slotter&);
    Slotter& operator = (const Slotter&);

};

