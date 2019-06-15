import * as React from 'react';
import { useReducer } from 'react';

export interface RootState {
  gameId?: string;
  players: string[];
}

const initialState: RootState = {
  players: [],
};

export const Store = React.createContext<{
  state: RootState;
  dispatch: React.Dispatch<any>;
}>({
  state: initialState,
  dispatch: value => {
    console.warn('State context not set up yet');
  },
});

function reducer(state: RootState, action: any) {
  switch (action.type) {
    case 'DO_SOMETHING': {
      return {
        ...state,
        gameId: 'hoorah!',
      };
    }
    default: {
      return state;
    }
  }
}

export const StoreProvider = (props: { children: React.ReactNode }) => {
  const [state, dispatch] = useReducer(reducer, initialState);
  const value = {
    state,
    dispatch,
  };

  return <Store.Provider value={value}>{props.children}</Store.Provider>;
};
