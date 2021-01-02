## How Authentication Works In The Blackspace Platform
---

Authentication is a very unique and distributed operations that involves a multitude of services. This document
aims to serve as a source of truth and further provide developers with a hollistic view of the operation.

Our authentication and authorization schemes make extensive use of the JWT tokens. The following set of operations will be
further elaborated upon in the below subsections.
* Registration (vanilla)
* Registration (OAUTH)
* Login (vanilla)
* Login (OAUTH)
* Logout
* Password Reset/Change

### Blackspace Registration
The registration flows varies across user types. Registration for business accounts and shoppers differs but from the context of the backend, the
level of interactions across services is similar.

### Registration (Vanilla)
Registration begins in the frontend and is a 2 step flow.
- through a form in the frontend, we collect a user's credentials and any other piece of data we require
- validate everything in the frontend and ensure all required fields are present
- send the information to the api gateway
    - api gateway obtain from the request object the email and password fields and calls the authentication_handler_service
    - authentication_handler_service performs a create account call via authentication service and returns the id of the created account to the
      api gateway
        - NOTE: authentication_handler_service must encrypt all passwords
      - if the gateway gets an error code instead of the id registration fails immediately and we return a response to the frontend
    - with the id, the api gateway then calls a set of backend services to create the account record and any additional necessary records
    /operations such as sending out an email validation email .... etc after performing validation.
      - if this fails, the gateway needs to be smart about retrying this operation. ensuring fields are properly validated before the request is
       sent from the frontend to the gateway can limit this from happening.
 - upon successful creation of the account, we send the response to the frontend.

### Registration (OAUTH)

### Login (Vanilla)
When a user logs in with blackspace, they establish two sessions: one with the frontend expires periodically, and another with the backend that can be
 used to refresh the app session. These are called the access token and refresh token, respectively.

During login, the backend works to ensure that users may not enumerate users in your system. This means it will not declare which field was incorrect
, but instead fails with a generic credentials error.
- through a form in the frontend, we collect user credentials and validate all fields
- send the information to the api gateway
    - gateway calls the authentication_handler_service which calls the authentication service to ensure it witholds the record of interest
        - if this fails we return an error to the frontend immediately
    - if the above call is successful, an account id would be returned. With this, we call a set of backend services dependent on wether the
     account is a business or shopping account to create a response.
     - if this fails we return an error immediately.
     - if this is successful, we return the data to the frontend

### Login (OAUTH)

### Logout
Most of our users will probably disappear when they close the application, but sometimes they will want to cleanly log out. The backend will take
 care of cleaning up its sessions (the the access token and refresh token).
 - When a user clicks the logout button, we call the api gateway
 - api gateway through the account id, calls the authentication_service_handler which calls the authentication service and invokes the logout api
    - the token within the session established with the backend for the account is revoked and the session is terminated
 - upon successful response, gateway calls any other cleanup operations through other microservices
 - return a response to the frontend or an error if any is encountered


 ### Password Reset/Change
The password registration process is a 2 phased endeavor. It requires the account owner to first request a password which will result in the
 account owner obtaining an email with a link to a password reset page where the password will be reset

 To request a password reset, a client will fill out a form containing their email.
 - frontend invokes the api gateway with the email field
 - api gateway calls the authentication_handler_service which calls the authentication_service via the /password/reset api invocation.
   the authentication service will post a webhook to the application password reset url with a request body containing the account_id and the jwt
    token. This webhook should be the url of the email service which should send an email containing the data to the user
   NOTE: APP_PASSWORD_RESET_URL must be set
- api gateway returns to the user either a 200 code or an error if anything occurs

From the password reset page, the frontend will submit the new password and the token to the gateway.
- gateway calls the authentication_handler_service which calls the authentication service via /password with the new password and the jwt token.
    - NOTE: authentication_handler_service must encrypt all passwords
- gateway calls and updates other services that must be aware of the new password such as sending a confirmation email that password was
 successfully change and returns the status of the operations to the frontend and any error if such occurs
