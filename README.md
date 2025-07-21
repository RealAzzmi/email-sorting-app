Warning: The Gmail API is somewhat slow, even when fetching the maximum 500 emails per request.

---

* You can use email protocols (IMAP, POP3, etc.) or the Gmail API (used here). With the Gmail API, you can opt for Pub/Sub for syncing emails (the recommended method for this challenge, paired with a message queue like RabbitMQ, which I didn't implement due to time constraints) or the History API (used here).
* A NoSQL document-based database would be more suitable here due to the document-based nature of emails and minimal joining, but I chose PostgreSQL for code simplicity.
* In a clean architecture, specific authentication tools like OAuth shouldn't reside in the domain layer, but I placed it there for simplicity.
* Several non-trivial optimizations could be made, such as:
    a. Caching
    b. Running background services
    c. Batching AI API calls and fully parallelizing API calls (the LLM model used is at 2000 RPM)
    d. Avoiding 'LIMIT OFFSET' in SQL queries, as cursor pagination (using an index with a `WHERE index > ...` clause) is more efficient.
* I was ill, so some aspects could be improved, such as better UX choices like custom-styled popups, auto-refreshing, and "select all pages" functionality.

---

### Extra Features Implemented

* The app's categories integrate with Gmail labels. This means if a user stops using the app, they can still view their emails organized by these categories in Gmail.