import React, { useState } from 'react';
import './App.css';

const aboba = '{"order_uid":"jeeeuid","track_number":"xmgpcqpgoftrack","entry":"WBIL","delivery":{"name":"RyanSidorov","phone":"+79854331467","zip":153151,"city":"Kyiv","address":"Stalina 33","region":"Solar system/Earth","email":"i0czhm13@gmail.com"},"payment":{"transactrion":"","request_id":"req_id","currency":"USD","provider":"wbpay","amount":1893,"payment_dt":8212686567,"bank":"Betta","delivery_cost":1036,"goods_total":857,"custom_fee":0},"items":[{"chrt_id":3054041,"track_number":"xmgpcqpgoftrack","price":24642,"rid":"ohx3pjwp93n6lg3","name":"ItemName","sale":0,"size":"0","total_price":24642,"nm_id":6935187,"brand":"ItemBrand","status":202}],"locale":"en","internal_signature":"intern_sign","customer_id":"RyanParker7335","delivery_service":"delivery","shardkey":"shardkey","Sm_id":12,"date_created":"DATA","oof_shard":"abc"}'
function Form({optionMsg, optionErr}) {
  const [value, setValue] = useState("1")
  const [nRequests, setNumbrRequests] = useState(0)
  let aboba2 = JSON.parse(aboba)
  //let a = aboba2.map(b => (<li>{b}</li>))
  function testFunc(){
    setNumbrRequests(nRequests + 1)
    console.log(nRequests)
    optionMsg(aboba2)
    if (nRequests % 2 == 0) {
      optionErr(true)
    } else {
      optionErr(false)
    }
    
  }
  async function postData(){
    console.log("button pressed")
      await Promise.all(fetch('http://localhost:8080', {
        method: 'POST',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: JSON.stringify({
          id: value
        })
      }).then(response => response.json())
      .then((resp) => (optionErr(resp.error), optionMsg(JSON.parse(resp.body)), console.log(resp.body), console.log(resp.error)))
      .then(setTimeout(() => {console.log("!")}, 1000)));
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
  setTimeout(() => { console.log("World!"); }, 2000);
  if (err === false){
    let item = message.items[0]
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
            <tr><td>ORDER_UID</td><td>{message.order_uid}</td></tr>
            <tr><td>TRACK_NUMBER</td><td>{message.track_number}</td></tr>
            <tr><td>ENTRY</td><td>{message.entry}</td></tr>
            <tr><td>LOCALE</td><td>{message.locale}</td></tr>
            <tr><td>INTERNAL_SIGNATURE</td><td>{message.internal_signature}</td></tr>
            <tr><td>CUSTOMER_ID</td><td>{message.customer_id}</td></tr>
            <tr><td>DELIVERY_SERVICE</td><td>{message.delivery_service}</td></tr>
            <tr><td>SHARDKEY</td><td>{message.shardkey}</td></tr>
            <tr><td>SM_ID</td><td>{message.sm_id}</td></tr>
            <tr><td>DATE_CREATED</td><td>{message.date_created}</td></tr>
            <tr><td>OOF_SHARD</td><td>{message.oof_shard}</td></tr>

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
            <tr><td>transaction</td><td>{message.payment.transaction}</td></tr>
            <tr><td>request_id</td><td>{message.payment.request_id}</td></tr>
            <tr><td>currency</td><td>{message.payment.currency}</td></tr>
            <tr><td>provider</td><td>{message.payment.provider}</td></tr>
            <tr><td>amount</td><td>{message.payment.amount}</td></tr>
            <tr><td>payment_dt</td><td>{message.payment.payment_dt}</td></tr>
            <tr><td>bank</td><td>{message.payment.bank}</td></tr>
            <tr><td>delivery_cost</td><td>{message.payment.delivery_cost}</td></tr>
            <tr><td>goods_total</td><td>{message.payment.goods_total}</td></tr>
            <tr><td>custom_fee</td><td>{message.payment.custom_fee}</td></tr>
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
            <tr><td>name</td><td>{message.delivery.name}</td></tr>
            <tr><td>phone</td><td>{message.delivery.phone}</td></tr>
            <tr><td>zip</td><td>{message.delivery.zip}</td></tr>
            <tr><td>city</td><td>{message.delivery.city}</td></tr>
            <tr><td>address</td><td>{message.delivery.address}</td></tr>
            <tr><td>region</td><td>{message.delivery.region}</td></tr>
            <tr><td>email</td><td>{message.delivery.email}</td></tr>
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
  } else if (err === true){
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
  const [message, setMsg] = useState()
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
