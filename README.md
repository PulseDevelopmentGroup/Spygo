# Spygo
Spyfall written in Go + React

## Websockets Schema

### Messages to server

#### Create Game

```json
{
  "type":"createGame",
  "data":{
    "username":"jsmith",
  }
}
```

`code`: Optional. If specified, a game with the given code will be created.

#### Destroy Game

```json
{
  "type":"destroyGame"
}
```

#### Join Game

```json
{
  "type":"joinGame",
  "data": {
    "code": "abc123",
    "username": "jsmith"
  }
}
```

`code`: Required. The code of the game to join
`username`: Required. The your username.

#### Leave Game

```json
{
  "type":"leaveGame"
}
```

### Messages to Client

#### Create Game Response

```json
{
  "type":"createGame",
  "data": {
    "sucess": true,
    "code":"abc123",
    "error": "..."
  }
}
```

`created`: Always returned. Whether or not the game was sucessfully created.
`code`: Returned if `created: true`. The code of the game created.
`error`: Returned if `created: false`. The error if the game was not created.

#### Destroy Game Response

```json
{
  "type":"destroyGame",
  "data":{
    "sucess": true,
    "error": "..."
  }
}
```

`destroyed`: Always returned. Whether or not the game was sucessfully destroyed.
`error`: Returned if `destroyed: false`. The error if the game was not destroyed.

#### Join Game Response

```json
{
  "type":"joinGame",
  "data":{
    "sucess": true,
    "error": "..."
  }
}
```

`joined`: Always returned. Whether or not the client joined the game sucessfully.
`error`: Returned if `joined: false`. The error if the client was not able to join the game.

#### Leave Game Response

```json
{
  "type":"leaveGame",
  "data":{
    "sucess": true,
    "error": "..."
  }
}
```

`left`: Always returned. Whether or not the client left the game sucessfully.
`error`: Returned if `left: false`. The error if the client was not able to leave the game.