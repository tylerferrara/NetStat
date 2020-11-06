import './App.css';
import React, {useState} from 'react';
import { withRouter , Route, Switch, Redirect} from 'react-router-dom';

import Home from "./Components/Login.js";
import Vote from "./Components/Vote.js";
import {getVotes, isAllowed} from "./Components/Auth.js";
import ProtectedRoute from "./Components/ProtectedRoute.js"; 


class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
        ssn: '',
        dob: '',
        isAuth: ''
    };
    this.setUserSSN = this.setUserSSN.bind(this);
    this.setUserDOB = this.setUserDOB.bind(this);
    this.authenticateUser = this.authenticateUser.bind(this); 
    this.handleLogin= this.handleLogin.bind(this); 
  }
  
  setUserSSN = (newSSN) => {
    this.setState({ssn: newSSN}); 
  };

  setUserDOB= (newDOB) => {
    this.setState({dob: newDOB}); 
  };

  setUserAuth= () => {
    this.setState({isAuth: true}); 
  };


  authenticateUser = async (ssn, dob) => {
    await this.setUserSSN(ssn); 
    await this.setUserDOB(dob); 

    let yes = await isAllowed(this.state.ssn, this.state.dob); 
    
    if (yes){
      await this.setUserAuth(); 
      return true; 
    }
    else return false; 
  }


  handleLogin = (e) => {
    e.preventDefault();
    this.props.history.push({
      pathname: '/vote',
      state: { ssn: this.state.ssn, dob: this.state.dob },
    }); 
  };


render() {
  return (
      <main>
            <Switch>
              <Route exact path="/" render={props => (<Home {...props} authenticateUser ={this.authenticateUser} getVotes={getVotes} handleLogin = {this.handleLogin}/>)}></Route>
              <ProtectedRoute path="/vote" component={Vote} ssn = {this.state.ssn} dob = {this.state.dob} isAuth = {this.state.isAuth}/>
            </Switch>
      </main>
  )
  }
}
export default withRouter(App);
// <Route path="/vote" render={props => (<Vote {...props} ssn = {this.state.ssn} dob = {this.state.dob} />)}></Route>
