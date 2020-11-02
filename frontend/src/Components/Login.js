import React from "react";
import '../App.css';
import {PostData} from '../serviceCalls.js';

class Login extends React.Component {

    constructor() {
        super();
        this.state = {
            ssn: '', 
            dob: '',
            isAuth: false
        };
  
        this.handleSSNChange = this.handleSSNChange.bind(this);
        this.handleDOBChange = this.handleDOBChange.bind(this);
        this.dismissError = this.dismissError.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.pleaseLetMeIn = this.pleaseLetMeIn.bind(this);
        this.pleaseRegisterMe = this.pleaseRegisterMe.bind(this); 
    }


    async pleaseLetMeIn(ssn, dob) { 
        //let's set up a response JSON 
        console.log("---------------------- CHECKING LOG IN ----------------------");
        let path = "login"; 
        let myData = {'SSN': ssn, 'DOB': dob};
        console.log(myData); 
    
        await PostData(path,myData).then((result) => {
            let responseJson = result;
            //if the backend says we can log in
            if(responseJson.Eligible){
                console.log("--------Eligible!-----------"); 
                let resp = {code: 200, message: "Valid User"}; 
                this.state.Auth = true; 
                return; 
            }
            else { 
                console.log (responseJson); 
                let resp = {code: 400, message: "Invalid User"}; 
                this.state.Auth = false; 
                return;  
            }
        });
               
    }
    
    //If we are not registered and want to vote, we check on register
    async pleaseRegisterMe(ssn, dob) { 
        console.log("---------------------- CHECKING REGISTER ----------------------");
        let path = "register"; 
        let myData = {'SSN': ssn, 'DOB': dob};
        console.log(myData); 
    
        await PostData(path,myData).then((result) => {
            let responseJson = result;
            //if the backend says we can log in
            if(responseJson.Success){
                console.log("--------Success!-----------"); 
                let resp = {code: 200, message: "Can be logged in!"}; 
                this.state.Auth = true; 
                return; 
            }
            else { 
                console.log (responseJson); 
                let resp = {code: 400, message: responseJson.Message}; 
                this.state.Auth = false; 
                return;  
            }
        });
    }
  
    handleSubmit(evt) {
        evt.preventDefault();
        console.log("Clicked!"); 

        if (!this.state.ssn) {
            return this.setState({ error: 'SSN is required' });
        }

        if (!this.state.dob) {
            return this.setState({ error: 'Date of Birth is required' });
        }

        this.pleaseLetMeIn(this.state.ssn, this.state.dob); 
        console.log(this.state.isAuth); 
        
        if (this.state.isAuth){
            console.log("We are done!");
            window.localStorage.setItem("Auth", true); 
            return this.setState({ error: '' }); 
        }

        else if (!this.state.isAuth){

            this.pleaseRegisterMe(this.state.ssn, this.state.dob); 
            console.log(this.state.isAuth); 

            if (this.state.isAuth){

                this.pleaseLetMeIn(this.state.ssn, this.state.dob);
                console.log(this.state.isAuth); 

                if (this.state.isAuth){
                    window.localStorage.setItem("Auth", true); 
                    return this.setState({ error: '' });
                }
                else if (!this.state.isAuth) {
                    window.localStorage.setItem("Auth", false);
                    return this.setState({ error: 'Failed Auth!' }); 
                }
            }
            else if (!this.state.isAuth){
                window.localStorage.setItem("Auth", false);
                return this.setState({ error: 'Failed Auth!' }); 
            }
        }

        return this.setState({ error: '' });
    }
  
    handleSSNChange(evt) {
        this.setState({
            ssn: evt.target.value,
        });
    };
  
    handleDOBChange(evt) {
        this.setState({
            dob: evt.target.value,
        });
    }

    dismissError() {
        this.setState({ error: '' });
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
            
                    <label>SSN: </label>
                    <input type="text" data-test="ssn" value={this.state.ssn} onChange={this.handleSSNChange} />

                    <label>DOB: </label>
                    <input type="text" data-test="dob" value={this.state.dob} onChange={this.handleDOBChange} />
        
                    <input type="submit" value="Log In" data-test="submit" />
                </form>
            </div>
      );
    }
  }
  
  export default Login;
