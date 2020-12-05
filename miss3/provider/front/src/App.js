import logo from './logo.svg';
import './App.css';
import React from 'react';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link
} from "react-router-dom";

const params = {}
const p = new URLSearchParams(window.location.search)
for (const param of p) {
  params[param[0]] = param[1]
}
console.log(params)

function App() {
  return (
    <Router>
      <Switch>
        <Route exact path="/">
          <Login />
        </Route>
      </Switch>
    </Router>
  );
}



class Login extends React.Component {
  constructor() {
    super()
    this.state = {
      username: "",
      password: "",
    }
  } 
  handleLogin(){
    console.log("here")
    console.log(this.state.username)
  }
  handleUserChange(e) {
    this.setState({username: e.target.value})
  }
  handlePassChange(e) {
    this.setState({password: e.target.value})
  }
  render() {
    return (
      <div>
        <h1>Secure Login</h1>
        <input placeholder="username" value={this.state.username} onChange={this.handleUserChange}/>
        <input placeholder="password" type="password" value={this.state.password} onChange={this.handlePassChange} />
        <button onClick={() => this.handleLogin()}>Sumbit</button>
      </div>
    )
  }

}

export default App;
