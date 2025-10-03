The purpose of this document is to provide a reference to all possible log
messages in the IVXV backend.

This document is partially auto-generated from the comments in the source code.
This is necessary to keep the document up to date through the evolution of the
software. The document is written in english, since the source code is written
in english for improved readability and availability to wider audiences.


IVXV systematically logs based on the protocol description and the service
status diagram. We describe general principles that are followed.

At a minimum, the following should be logged:

* The fact of receiving each request and the start of processing

* The handover of processing to an external component

* The return of the processing sequence to the component

* The end of the request processing and the result

* Additionally, passing through significant stages in the process state model together with important facts about the process


General Principles:

* The rsyslog service is used for logging, which records the timestamp of the log message with millisecond precision.

* At the start of each session, the system generates a unique identifier, which the client application uses in its requests to the central system.

* All log entries belonging to a single session contain the same session identifier. Validity of the session identifier is checked, requests with unknown, invalid or expired session identifiers are not served.

* Each log entry must be uniquely identifiable.

* For each log message, it must be possible to unambiguously identify the location in the source code where the message originated. To ensure this, each logged message must have a unique phrasing or contain a unique identifier. In IVXV all event identifiers and corresponding data structures are generated before the compilation. This makes it easier to follow unified pattern.

* The log entry should preferably be in JSON format, where machine readability is primary and human readability is secondary for automated monitoring.

* The log entries may be stacked, to include information from more technical contexts.

* If in doubt - log, one can filter unnecessary information, but cannot regenerate missing information.

Requirements:

* The information logged must be sanitized (e.g., URL encoding) and subject to length restrictions (both overall and per parameter).

* Information originating from outside the system perimeter must only be logged in a sanitized form and within a predefined length.

* Certificates and the agreed cipher used during TLS handshake must be logged.

* Sensitive data must not be logged Directly, it must be protected by e.g. hashing.
