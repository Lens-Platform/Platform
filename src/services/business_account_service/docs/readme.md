## Business Account Service

#### Table Of Contents
- [Business Account Service](#business-account-service)
    - [Table Of Contents](#table-of-contents)
    - [Overview](#overview)
    - [Dependencies](#dependencies)
    - [Service Level Interactions (SLAs)](#service-level-interactions-slas)
        - [Account Sign Up Flow](#account-sign-up-flow)
        - [Account Login Flow](#account-login-flow)
        - [Account Archive Flow](#account-archive-flow)
    - [Testing](#testing)
        - [Stress & Chaos Testing](#stress--chaos-testing)

#### Overview
The business account service manages all interactions and features specific to our merchant/business profiles. Some interactions that this service
 handles include, authentication and authorization of business accounts, CRUD operations against business accounts, as well as business acount
  onboarding amongst many other features.

#### Dependencies
This service witholds strict dependencies on the authentication handler service as well as the api-gateway. It leverages the authentication service
 to perform distributed account operations some of which include account locking, unlocking, archiving, ...etc

#### Service Level Interactions (SLAs)
##### Account Sign Up Flow
The account sign up process starts at the client. The below steps outline the set of service level interactions that take place during the account sign up process.
1. API Gateway obtains client request with sign up data. API Gateway validates this data.
   1. The API gateway initiates a saga comprised of 2 steps.
      1. Step 1: invokes `/signup` api of the authentication handler service.
         1. On failure we abort the saga and return to the client.
         2. On success we continue to the next step
      2. Step 2: depending on wether the sign up api was triggered by a shopper or merchant, we trigger the sign up flow of the business account service or the shopper service.
         1. On failure, we call the authentication handler service and attempt to delete the created record. If this fails, we place the request onto a queue for a worker job to perform the delete operation at a later time. If the worker job, performs the request and it too fails, we re-emplace the request metadata into the queue
         2. On success, we return a success status to the client

This operation requires interaction with the authentication handler service, the authentication service, the business account or shopper service as well as a queue through which a worker job will process the request.

It is important that we have the proper alerts in place to trigger when a service replica becomes unhealthy, when the queue becomes full/overwhelmed, or requests at either interacting service commence to fail.
##### Account Login Flow
The account login process starts at the client. The below steps outline the set of service level interactions that take place during the login process.
1. API Gateway obtains client request with log in data. API Gateway validates this data
   1. The API Gateway initiates a saga comprised of a multitude of steps
      1. Step 1: invokes the `login` api endpoint of the authentication handler service
         1. On failure we abort the operation and return to the client
         2. On success we obtain a JWT token.
      2. Step 2: dependent on wether the client provided a request on the behalf of a shopper or business account, we place the JWT token in the request header at the API Gateway layer and invoked the necessary set of services whose responses we return to the client (good or bad)

This operation requires interaction with the authentication handler service, the authentication service, and any other set of dependent services

The proper alerting schemes must be set in place, not to mention metrics and request tracing plus logs.
##### Account Archive Flow
The account login process starts at the client. The below steps outline the set of service level interactions that take place during the login process.
1. API Gateway obtains client request with log in data. API Gateway validates this data
   1. The API Gateway initiates a sage comprised of a multitude of steps
      1. Step 1: dependent on wether the operation is generated on the behalf of a merchant or shopper account, we invoke the archive api of the appropriate service. The archive api of any necessary service (business account and shopper) performs the archive operation in a saga comprised of 2 steps.
         1. In the first step the  authentication handler service's lock account api endpoint is invoked
            1. Upon failure, an api invocation is performed against the authentication handler service to unlock the account. This is performed in case the operation was actually successful on the behalf of the authentication handler service but the response was not received due to some reason (network issue, ...etc). The operation is then terminated and the response provided to the client. If this too fails we return to the client.
            2. On success the next step is performed
         2. In the second step, the account's activity status is set to inactive. If this fails, then account is reset to its inital state and the unlock api endpoint in the authentication handler service is invoked.
#### Testing
This codebase witholds both unit tests and service level integration testing.

To properly run all tests (integration and unit tests) it is imperative that all dependent services are up and running prior to the tests being launched.
##### Stress & Chaos Testing
Stress & Chaos testing is not yet supported but will be soon.
