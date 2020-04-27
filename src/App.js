
import React, {useState,useCallback} from 'react';
import Game from './components/Game/Game'
import UserLogin from './components/userLogin/UserLogin'
import './App.css';

function App() { 

  const [username, setUsername] = useState("")
  const [isGameReady, setIsGameReady] = useState(false)

  const changeInput =(e)=>{
    setUsername(e.target.value)
  }
  const onclick=useCallback(()=>{
    if(!username)return
    setIsGameReady(true)
  },[username])
  return (
    <>  
    {isGameReady ? <Game username={username} /> : <UserLogin val={username}  changeInput={changeInput} onclick={onclick}/>}
        
  </>
  );
}
export default App;