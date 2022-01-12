import React, { Component, useState } from 'react';
import './App.css';
function Form() {
  const [value, setValue] = useState("1")
  const [res, setRes] = useState()
  const [nRequests, setNumbrRequests] = useState(0)
  let aboba = 0
  function postData(){
    console.log("button pressed")
      fetch('http://localhost:8080', {
        method: 'POST',
        //mode: 'no-cors',
        //mode: origin = *,
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: JSON.stringify({
          id: value
        })
      }).then(response => response.json()).then((resp) => setRes(resp.value));
      // const data = await result.json();
      console.log("REQUESTED" + value);
      //console.log(res);
      setNumbrRequests(nRequests + 1)
    }
  if (nRequests > 0){
    console.log(res)
    console.log(nRequests)
  }
  //let val = "123"
  return(
  <div className='Form'>
    <button className='Form-button' onClick={ () => postData()}> Search</button>
    <input className='Form-input' value={value} onChange={(e) => setValue(e.target.value)}/>
  </div>)
}

function Result() {

}
function App() {
  return(
  <div className='App'>
    <header className='App-header'>Simple form to test MyApp
      <Form />
    </header>
    
  </div>)
}
export default App;
