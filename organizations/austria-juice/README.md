# End to end test Austria Juice

This e2e test is meant to prevent regression in the austria juice email pipeline in case changes are made to the mail server, the email pipeline or associated configuration.

## Test dependencies
As an e2e test, it uses the mail server on tnf-mail.unchain.io with a test email address (`austria-juice-staging@tnf-mail.unchain.io`).
The import API is mocked with a local http server that receives the POST message and can perform assertions on it.

A local SMTP server is started to check error messages. This email server is configured in the docker-compose file in the root directory.

## Test scenario: success
Test setup
1. A new email with the austria juice CSV attachment (`example.csv`) is send to the IMAP inbox of the above mentioned email.
2. A HTTP server is started on localhost:80
Test execution
3. The pipeline is triggered in order to fetch and process the email sent in step 1.
Test assertions
4. The incoming message on the local HTTP server is checked for the expected body that should be sent to the import API.
5. The IMAP inbox is checked to make sure the message is marked as read
6. The SMTP server is checked for not having error messages
Test teardown
5. The pipeline is stopped

# Test scenario: fail over
Test setup
1. A new email with an invalid CSV attachment (`bad_example.csv`) is send to the IMAP inbox of the above mentioned email.
2. A HTTP server is started on localhost:80
Test execution
3. The pipeline is triggered in order to fetch and process the email sent in step 1.
Test assertions
4. The IMAP failed mailbox is checked to make it contains a moved message
5. The SMTP server is checked for containing one error messages
Test teardown
6. The pipeline is stopped
