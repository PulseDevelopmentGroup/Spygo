import React, { Component, useState, useEffect, useContext } from 'react';
import { BrowserRouter as Router, Route, Link } from 'react-router-dom';
import styled from 'styled-components';

import { packMessage } from './utils/socketUtils';
import MessageBroker from './utils/messageBroker';

import GlobalStyles from './components/Global';
import Landing from './pages/Landing';
import Game from './pages/Game';
import { Store } from './store';

const Container = styled.div`
  display: flex;
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;

  flex-direction: column;
  justify-content: center;
  padding: 15px;

  background-color: #fdfdff;
`;

interface Props {}

const App: React.FC<Props> = (props: Props) => {
  const { state, dispatch } = useContext(Store);

  const [socket, setSocket] = useState<WebSocket | undefined>();
  const [gameData, setGameData] = useState({});

  const onSocketMessage = (message: any) => {
    console.log('Message recieved');
    console.log(message);
  };

  const createGame = (username: string) => {
    const gameObj = {
      username: username,
    };

    const payload = packMessage('CREATE_GAME', JSON.stringify(gameObj));
    socket && socket.send(payload);
  };

  const joinGame = (id: string, username: string) => {
    const gameObj = {
      gameId: id || '',
      username: username,
    };

    const payload = packMessage('JOIN_GAME', JSON.stringify(gameObj));
    socket && socket.send(payload);
  };

  const gameFunctions = {
    createGame,
    joinGame,
  };

  const setGame = (payload: any) => {
    const { gameId: id, username } = payload.data;

    setGameData({
      id,
      username,
    });
  };

  useEffect(() => {
    console.log(state, dispatch);
    dispatch({ type: 'DO_SOMETHING' });

    const socket = new WebSocket(`ws://${process.env.REACT_APP_API_URL}/api`);
    setSocket(socket);

    MessageBroker.subscribe('JOIN_GAME', setGame);

    socket.onopen = e => {
      let message = {
        type: 'create-game',
        data: '{"code":"", "username":"user"}',
      };

      let obj = JSON.stringify(message);

      socket.onerror = err => {
        console.log(`Following error occured with websocket: ${err}`);
      };

      socket.onmessage = e => {
        MessageBroker.handleMessage(e);
      };

      window.onbeforeunload = () => {
        console.log('Closing socket');
        socket.close();
      };
    };
  }, []);

  return (
    <Router>
      <Container>
        <GlobalStyles />
        <Link to="/game/abc123">This link</Link>
        <Route
          exact
          path="/"
          render={props => <Landing {...props} gameFunctions={gameFunctions} />}
        />
        <Route exact path="/game/:gameid" component={Game} />
      </Container>
    </Router>
  );
};

export default App;
