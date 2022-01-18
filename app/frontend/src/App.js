import React, { useState } from 'react';
import './App.css';

const aboba = '{"order_uid":"jeeeuid","track_number":"xmgpcqpgoftrack","entry":"WBIL","delivery":{"name":"RyanSidorov","phone":"+79854331467","zip":153151,"city":"Kyiv","address":"Stalina 33","region":"Solar system/Earth","email":"i0czhm13@gmail.com"},"payment":{"transactrion":"","request_id":"req_id","currency":"USD","provider":"wbpay","amount":1893,"payment_dt":8212686567,"bank":"Betta","delivery_cost":1036,"goods_total":857,"custom_fee":0},"items":[{"chrt_id":3054041,"track_number":"xmgpcqpgoftrack","price":24642,"rid":"ohx3pjwp93n6lg3","name":"ItemName","sale":0,"size":"0","total_price":24642,"nm_id":6935187,"brand":"ItemBrand","status":202}],"locale":"en","internal_signature":"intern_sign","customer_id":"RyanParker7335","delivery_service":"delivery","shardkey":"shardkey","Sm_id":12,"date_created":"DATA","oof_shard":"abc"}'
function Form({optionMsg, optionErr}) {
  const [value, setValue] = useState("1")
  const [nRequests, setNumbrRequests] = useState(0)
  let aboba2 = JSON.parse(aboba)
  //let a = aboba2.map(b => (<li>{b}</li>))
  function testFunc(){
    const mess = {
      body: '',
      error: false
    }
    setNumbrRequests(nRequests + 1)
    console.log(nRequests)
    let obj = Object.create(mess)
    obj.body = aboba
    
    if (nRequests % 2 == 0) {
      obj.error = true
      optionMsg(obj)
    } else {
      obj.error = false
      optionMsg(obj)
    }
    
  }
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
      }).then(response => response.json())
      .then((resp) => (optionMsg(resp), optionErr(resp.error), console.log(resp), console.log(resp)))
      .then(setTimeout(() => {console.log("!")}, 1000));
      console.log("REQUESTED" + value);
      //console.log(res);
      
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

function Content({message, err}) {
  console.log("from container")
  console.log(err)
  console.log(typeof(message))
  console.log(message)
  console.log(message.body)
  console.log(message.error)
  if (message.error === false){
    let msg = JSON.parse(message.body)
    let item = msg.items[0]
    console.log(aboba)
    //message = JSON.parse(message)
    return(
      <div>
        BASICS
        <table className='table'>
          <thead>
            <tr>
              <th>KEY</th><th>VALUES</th>
            </tr>
          </thead>
          <tbody>
            <tr><td>ORDER_UID</td><td>{msg.order_uid}</td></tr>
            <tr><td>TRACK_NUMBER</td><td>{msg.track_number}</td></tr>
            <tr><td>ENTRY</td><td>{msg.entry}</td></tr>
            <tr><td>LOCALE</td><td>{msg.locale}</td></tr>
            <tr><td>INTERNAL_SIGNATURE</td><td>{msg.internal_signature}</td></tr>
            <tr><td>CUSTOMER_ID</td><td>{msg.customer_id}</td></tr>
            <tr><td>DELIVERY_SERVICE</td><td>{msg.delivery_service}</td></tr>
            <tr><td>SHARDKEY</td><td>{msg.shardkey}</td></tr>
            <tr><td>SM_ID</td><td>{msg.sm_id}</td></tr>
            <tr><td>DATE_CREATED</td><td>{msg.date_created}</td></tr>
            <tr><td>OOF_SHARD</td><td>{msg.oof_shard}</td></tr>

          </tbody>
        </table>
        PAYMENT
        <table className='table'>
          <thead>
            <tr>
              <th>KEY</th><th>VALUES</th>
            </tr>
          </thead>
          <tbody>
            <tr><td>transaction</td><td>{msg.payment.transaction}</td></tr>
            <tr><td>request_id</td><td>{msg.payment.request_id}</td></tr>
            <tr><td>currency</td><td>{msg.payment.currency}</td></tr>
            <tr><td>provider</td><td>{msg.payment.provider}</td></tr>
            <tr><td>amount</td><td>{msg.payment.amount}</td></tr>
            <tr><td>payment_dt</td><td>{msg.payment.payment_dt}</td></tr>
            <tr><td>bank</td><td>{msg.payment.bank}</td></tr>
            <tr><td>delivery_cost</td><td>{msg.payment.delivery_cost}</td></tr>
            <tr><td>goods_total</td><td>{msg.payment.goods_total}</td></tr>
            <tr><td>custom_fee</td><td>{msg.payment.custom_fee}</td></tr>
          </tbody>
        </table>
        DELIVERY
        <table className='table'>
          <thead>
            <tr>
              <th>KEY</th><th>VALUES</th>
            </tr>
          </thead>
          <tbody>
            <tr><td>name</td><td>{msg.delivery.name}</td></tr>
            <tr><td>phone</td><td>{msg.delivery.phone}</td></tr>
            <tr><td>zip</td><td>{msg.delivery.zip}</td></tr>
            <tr><td>city</td><td>{msg.delivery.city}</td></tr>
            <tr><td>address</td><td>{msg.delivery.address}</td></tr>
            <tr><td>region</td><td>{msg.delivery.region}</td></tr>
            <tr><td>email</td><td>{msg.delivery.email}</td></tr>
          </tbody>
        </table>
        ITEM
        <table className='table'>
          <thead>
            <tr>
              <th>KEY</th><th>VALUES</th>
            </tr>
          </thead>
          <tbody>
            <tr><td>chrt_id</td><td>{item.chrt_id}</td></tr>
            <tr><td>track_number</td><td>{item.track_number}</td></tr>
            <tr><td>price</td><td>{item.price}</td></tr>
            <tr><td>rid</td><td>{item.rid}</td></tr>
            <tr><td>name</td><td>{item.name}</td></tr>
            <tr><td>sale</td><td>{item.sale}</td></tr>
            <tr><td>size</td><td>{item.size}</td></tr>
            <tr><td>total_price</td><td>{item.total_price}</td></tr>
            <tr><td>nm_id</td><td>{item.nm_id}</td></tr>
            <tr><td>brand</td><td>{item.brand}</td></tr>
            <tr><td>status</td><td>{item.status}</td></tr>
          </tbody>
        </table>
      </div>
    )
  } else if (message.error === true){
    //setOrder_uid("")
    return(
      <div>
        <p>THERE IS NO SUCH VALUE IN MY DATABASE</p>
      </div>
    )
  }
  return(<div>
    text
  </div>)
}
function App() {
  const [message, setMsg] = useState({})
  const [err, setErr] = useState()
  console.log(err)
  
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
      
    </header>
    <Content className='Data' message={message} err={err}/>
  </div>)
}
export default App;
