import React, { Component, useState } from 'react';
import './App.css';


function Form({optionMsg, optionErr}) {
  const [value, setValue] = useState("1")
  const [nRequests, setNumbrRequests] = useState(0)
  let aboba = 0
  function postData(){
    console.log("button pressed")
      fetch('http://localhost:8080', {
        method: 'POST',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: JSON.stringify({
          id: value
        })
      }).then(response => response.json()).then((resp) => (optionMsg(resp.body), optionErr(resp.err), console.log(resp.body), console.log(resp.error)));
      // const data = await result.json();
      console.log("REQUESTED" + value);
      //console.log(res);
      setNumbrRequests(nRequests + 1)
    }
  if (nRequests > 0){
    console.log(nRequests)
  }
  return(
  <div className='Form'>
    <button className='Form-button' onClick={ () => postData()}> Search</button>
    <input className='Form-input' value={value} onChange={(e) => setValue(e.target.value)}/>
  </div>)
}

// function Content() {
//   return(
//   <div >
//     {err}
//     <br>
//     </br>
//     {message}
//   </div>
//   )
// }
function App() {
  const [message, setMsg] = useState()
  const [err, setErr] = useState()

  const setFormMsg = (msg) => {
    setMsg(msg)
  }

  const setFormErr = (err) => {
    setErr(err)
  }
  return(
  <div className='App'>
    <header className='App-header'>Simple form to test MyApp
      <Form optionMsg={setFormMsg} optionErr={setFormErr}/>
      {message}
      {err}
    </header>
    
  </div>)
}
export default App;
