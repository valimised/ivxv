#pragma once

class Counter {

public:

    Counter();
    Counter(const Counter&);
    Counter& operator = (const Counter&);

    virtual ~Counter();

    void registerRequest(int size, bool success);

    int lastRequest;
    int lastSuccessRequest;
    int maxRequest;
    int maxSuccessRequest;
    int cntRequest;
    int cntSuccessRequest;

protected:

    void init(const Counter&);

};

