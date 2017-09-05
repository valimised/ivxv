#include "Counter.h"


Counter::Counter()
{
    lastRequest = 0;
    lastSuccessRequest = 0;
    maxRequest = 0;
    maxSuccessRequest = 0;
    cntRequest = 0;
    cntSuccessRequest = 0;
}

Counter::Counter(const Counter& other)
{
    init(other);
}

Counter& Counter::operator = (const Counter& other)
{
    if (&other != this) {
        init(other);
    }

    return *this;
}

void Counter::init(const Counter& other)
{
    lastRequest = other.lastRequest;
    lastSuccessRequest = other.lastSuccessRequest;
    maxRequest = other.maxRequest;
    maxSuccessRequest = other.maxSuccessRequest;
    cntRequest = other.cntRequest;
    cntSuccessRequest = other.cntSuccessRequest;
}

Counter::~Counter()
{
}

void Counter::registerRequest(int size, bool success)
{
    lastRequest = size;

    if (lastRequest > maxRequest) {
        maxRequest = lastRequest;
    }

    cntRequest += 1;

    if (success) {
        lastSuccessRequest = size;

        if (lastSuccessRequest > maxSuccessRequest) {
            maxSuccessRequest = lastSuccessRequest;
        }

        cntSuccessRequest += 1;
    }
}

