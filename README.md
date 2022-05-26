## Web Application API Template for user registration and login flow. 

This covers signup, login endpoints as well as an additional resource `users` that can be accessed if you are logged in.

To determine whether a request is from a logged in user or not, Json Web Tokens (https://jwt.io/) are used. The frontend will be sending requests with the JWT in the `x-authentication-token` header.

PostgreSQL is used as a database, in its own docker container

## API Specifications

### `POST /signup`
Endpoint to create a user row in the db. The payload should have the following fields:

```json
{
  "email": "email@mydomain.com",
  "password": "123test",
  "firstName": "Ed",
  "lastName": "Markentype"
}
```

where `email` is a unique key in the database.
The response body should return a JWT on success that can be used for other endpoints:

```json
{
  "token": "some_jwt_token" 
}
```

### `POST /login`
Endpoint to log a user in. The payload should have the following fields:

```json
{
  "email": "email@mydomain.com",
  "password": "123test"
}
```

The response body should return a JWT on success that can be used for other endpoints:

```json
{
  "token": "some_jwt_token"
}
```

### `GET /users`
Endpoint to retrieve a json of all users. This endpoint requires a valid `x-authentication-token` header to be passed in with the request.

The response body should look like:
```json
{
  "users": [
    {
      "email": "email@mydomain.com",
      "firstName": "Ed",
      "lastName": "Markentype"
    }
  ]
}
```

### `PUT /users`
Endpoint to update the current user `firstName` or `lastName` only. This endpoint requires a valid `x-auhentication-token` header to be passed in and it should only update the user of the JWT being passed in. The payload can have the following fields:

```json
{
  "firstName": "NewFirstName",
  "lastName": "NewLastName"
}
```

The response body can be empty.

## Startup
The project docker definition is in `docker-compose.yml`
everything should start with a simple

`docker-compose up`

which will bring up a database container (that in turns  will create a local `pgdata/` folder for the storage)

### first time database creation
`docker-compose run api app createDb`

tested with:
```
Docker version 20.10.12, build e91ed57
Docker Compose version v2.2.3
```

### known problems
> - password is not encrypted in the db
> - there is no way to expire tokens that are active (other than just waiting: the database could be used for this, making it possible to expire tokens even before their time is up) 
> - there is no protection or caching of the /signup endpoint, at the moment every request causes a database hit - this could be smarter to try and avoid DoS-like attempts
