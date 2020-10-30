import React from "react";
import '../App.css';
import {PostData} from '../serviceCalls.js';

class Login extends React.Component {

    constructor() {
        super();
        this.state = {
            username: '',
            password: '',
            error: '',
            loggedIn: false
        };
  
        this.handlePassChange = this.handlePassChange.bind(this);
        this.handleUserChange = this.handleUserChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.dismissError = this.dismissError.bind(this);
        this.pleaseLetMeIn = this.pleaseLetMeIn.bind(this);
        this.pleaseRegisterMe = this.pleaseRegisterMe.bind(this); 
    }
  
    dismissError() {
        this.setState({ error: '' });
    }
    //curl -v -XGET -H "Content-type: application/json" -d '{"SSN": "111110", "DOB":"12/10/1991"}' 'localhost:8080/login'

    //If we are already registered and want to log in, just log in
    pleaseLetMeIn(){
        console.log("Please let me in!");
        let path = "login"; 
        let myData = {'SSN': this.state.username, 'DOB': this.state.password};

        PostData(path,myData).then((result) => {
            let responseJson = result;
            if(responseJson.eligible == true){
                console.log("We're logged in!"); 
                this.setState({loggedIn: true}); 
                return true; 
            }
            else { return false; }
            //if not, set error message
        });
    }

    //If we are not registered and want to vote, we check on register
    pleaseRegisterMe(){
        console.log("Please register me!");
        let path = "register"; 
        let myData = {'SSN': this.state.username, 'DOB': this.state.password};

        PostData(path,myData).then((result) => {
            let responseJson = result;
            if(responseJson.success){
                console.log("We're registered!"); 
                //now we log in 
                this.pleaseLetMeIn; 
                return true; 
            }
            //if not, set error message
            else {return false;}
        });
    }
  
    handleSubmit(evt) {
        evt.preventDefault();
  
        if (!this.state.username) {
            return this.setState({ error: 'Username is required' });
        }
  
        if (!this.state.password) {
            return this.setState({ error: 'Password is required' });
        }

        if (!this.pleaseLetMeIn) {
            console.log("This user is not registered!"); 
            if (! this.pleaseRegisterMe) {
                console.log("This user is not eligible!"); 
                return this.setState({ error: 'This user is not eligible!' });
            }
        }
        return this.setState({ error: '' });
    }
  
    handleUserChange(evt) {
        this.setState({
            username: evt.target.value,
        });
    };
  
    handlePassChange(evt) {
        this.setState({
        password: evt.target.value,
        });
    }
  
    render() {
        return (
            <div className="Login">
                <form onSubmit={this.handleSubmit}>
                {
                    this.state.error &&
                    <h3 data-test="error" onClick={this.dismissError}>
                        <button onClick={this.dismissError}>✖</button>
                    {this.state.error}
                    </h3>
                }   
            
            <label>User Name</label>
            <input type="text" data-test="username" value={this.state.username} onChange={this.handleUserChange} />

            <label>Password</label>
            <input type="password" data-test="password" value={this.state.password} onChange={this.handlePassChange} />
  
            <input type="submit" value="Log In" data-test="submit" />
          </form>
        </div>
      );
    }
  }
  
  export default Login;
