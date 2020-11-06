import React from "react";
import '../App.css';

class Login extends React.Component {

    constructor() {
        super();

        this.state = {
            ssn: '',
            dob: '',
            error: '',
            isAuth: false,
            votePayload: {
                "Minushka": 0,
                "Zach": 0,
                "Final": false,
            }
        }
  
        this.handleSSNChange = this.handleSSNChange.bind(this);
        this.handleDOBChange = this.handleDOBChange.bind(this);
        this.dismissError = this.dismissError.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    componentDidMount() {
        this.props.getVotes(result => {
            this.setState({votePayload: result});
        })
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

    handleSubmit = async (e) => {
        e.preventDefault();
        await this.props.authenticateUser(this.state.ssn, this.state.dob); 
        this.props.handleLogin(e); 
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
                <div>
                    <h3>{ (this.state.votePayload.Final ? "Final " : "Ongoing ")}Election Votes</h3>
                    <h5>{ "Minushka: " + this.state.votePayload.Minushka + "\n" }</h5>
                    <h5>{ "Zach: " + this.state.votePayload.Zach + "\n" }</h5>
                </div>
            </div>
      );
    }
  }
  
  export default Login;
