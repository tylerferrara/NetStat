import React from "react";
import '../App.css';

class HomePage extends React.Component {

    constructor() {
        super();
        this.state = {
            username: '',
            password: '',
            error: '',
        };
  
        this.handlePassChange = this.handlePassChange.bind(this);
        this.handleUserChange = this.handleUserChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.dismissError = this.dismissError.bind(this);
    }
  
    dismissError() {
        this.setState({ error: '' });
    }
  
    handleSubmit(evt) {
        evt.preventDefault();
  
        if (!this.state.username) {
            return this.setState({ error: 'Username is required' });
        }
  
        if (!this.state.password) {
            return this.setState({ error: 'Password is required' });
        }

        /* Send username and password to backend */
        var payload = {"SSN": this.state.username, "DOB": this.state.password};
        fetch(
            'http://localhost:8080/login', {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json;charset=utf-8'
                },
                body: JSON.stringify(payload)
        }).then(data => data.json())
        .then(result => {
            console.log(result)
            if (result != null && result.Eligible) {
                console.log("YAY!!")
            }
        })
        .catch(err => console.log(err))

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
  
  export default HomePage;
