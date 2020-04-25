import { useReducer, useEffect, useRef, useCallback } from 'react';

const initialState = {
  exceptions: [],
  players: []
};
const gameReducer = (state, action) => {
  switch (action.type) {
    case "player":
      const { player } = action.by
      const objIndex = state.players.findIndex(obj => obj.id === player.id);
      if (objIndex !== -1) {
        const clonePlayers = JSON.parse(JSON.stringify(state.players));
        //check if player need to be seen in the screen if not remove from players
        if (!player.show) {
          clonePlayers.splice(objIndex, 1);
          return {
            ...state,
            players: clonePlayers
          }
        }
        //check if player hit bounderies and add animation
        player.collision ? clonePlayers[objIndex].shake = true : clonePlayers[objIndex].shake = false

        //change player cordinates on screen
        clonePlayers[objIndex].x = player.x;
        clonePlayers[objIndex].y = player.y;
        clonePlayers[objIndex].score = player.score;
        return {
          ...state,
          players: clonePlayers
        }
      }
      return {
        ...state,
        players: [
          ...state.players,
          player
        ]
      };
    case "exception":
      const { exception } = action.by
      const exceptionIndex = state.exceptions.findIndex(obj => obj.exceptionType === exception.exceptionType);
     
      if (exceptionIndex !== -1) {
        const cloneExceptions = JSON.parse(JSON.stringify(state.exceptions));
        if (!exception.show) {
          cloneExceptions.splice(exceptionIndex, 1);
          return {
            ...state,
            exceptions: cloneExceptions
          }
        }
      }
      if(!exception.show) return {...state}
      return {
        ...state,
        exceptions: [...state.exceptions, exception]
      };
    case "self":
      const { self } = action.by
      let newPlayer = { ...self, active: true }
      return {
        ...state,
        players: [
          ...state.players,
          newPlayer
        ]
      };
    default:
      throw new Error();
  }
};

export const useWebSocket = (url, bounderies) => {
  const [messages, dispatch] = useReducer(gameReducer, initialState);
  const webSocket = useRef(null);

  useEffect(() => {
    webSocket.current = new WebSocket(url);
    webSocket.current.onmessage = (event) => {
      const parseData = JSON.parse(event.data);
      const [wsInfo] = Object.getOwnPropertyNames(parseData)
      dispatch({ type: wsInfo, by: parseData });
    };

  }, [url]);


  useEffect(() => {
    webSocket.current.onopen = () => {
      webSocket.current.send(JSON.stringify(bounderies.current))
    }
    return () => {
      webSocket.current.onclose = (e) => {
        console.log('e',e)
      }
    };
  }, [bounderies]);

  const sendMessage = useCallback(message => {
    if (!message) return
    webSocket.current.send(JSON.stringify(message));
  }, [webSocket]);

  return [messages, sendMessage,webSocket.current]
};
