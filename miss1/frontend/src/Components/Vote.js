import React from 'react';
import Button from 'react-bootstrap/Button'
import '../App.css';
import {PostData} from '../serviceCalls.js';
import {withRouter} from 'react-router-dom'

import Minush from '../Assets/Minushka.jpg'
import Zach from '../Assets/Zach.jpg'

 
class Vote extends React.Component {
    constructor(props){
      super(props);
      this.state = {
        ssn: this.props.history.location.state.ssn,
        dob: this.props.history.location.state.dob,
        candidateClicked: -1,
        isConfirmed: false, 
        hasVoted: false,
        error: ''
      };

      this.whichCandidateDidIVoteFor = this.whichCandidateDidIVoteFor.bind(this);
      this.submittingMyVote = this.submittingMyVote.bind(this);
  }
  
    //hardcoded portion to return candidate string
    whichCandidateDidIVoteFor() {
      if (this.state.candidateClicked == 0 ){
        return "Minushka";
      }
      if (this.state.candidateClicked == 1 ){
        return "Zach";
      }
      else return ""; 
    }

    //Changes the candidate voted for when clicked
    minushClick = () => {
      this.setState({
        candidateClicked : 0,
      })
    }    
    
    //Changes the candidate voted for when clicked
    zachClick = () => {
        this.setState({
          candidateClicked : 1,
        })
    } 

    //Sends the REST call to submit vote
    async submittingMyVote() {
      let path = "vote"; 
      let myData = {"SSN": this.state.ssn, "DOB": this.state.dob, "Candidate": this.whichCandidateDidIVoteFor()};

      await PostData(path,myData).then((result) => {
        let responseJson = result;
        if(responseJson.Success == true){
            console.log (responseJson);
            this.setState({ error: responseJson.Message});
            this.setState({
              isConfirmed : true,
              hasVoted: true
            })
        }
        else { 
            console.log (responseJson); 
            this.setState({ error: responseJson.Message});
            return false; }
      });
    }


    confirmSubmission = async () => {
        if (this.state.isDone && this.state.isConfirmed){
          console.log("Already voted!"); 
          return; 
        }
        console.log("Gonna vote!"); 
        await this.submittingMyVote(); 
    } 
  
    render () {
      return (
        <div>
            <div>
              <div>User is: {this.state.ssn} and  {this.state.dob}. {this.state.error}</div>
                <img src={Minush} onClick={this.minushClick} width = "300"/>
                {
                  
                    (this.state.candidateClicked ==0 && !this.state.isDone && !this.state.hasVoted) && //only show if we are allowed to vote
                    <div>You clicked Minushka!
                        <Button onClick={this.confirmSubmission}> Confirm </Button> {
                            this.state.isConfirmed && !this.state.isDone && !this.state.hasVoted  &&<div> "You're Confirmed for Minushka!" {this.state.error}</div>
                        }
                    </div>
                }
            </div>
            <div>
                <img src={Zach} onClick={this.zachClick} width = "300"/>
                {
                    (this.state.candidateClicked ==1 && !this.state.isDone && !this.state.hasVoted) && //only show if we are allowed to vote
                    <div>You clicked Zach!
                        <Button onClick={this.confirmSubmission}> Confirm </Button> {
                            this.state.isConfirmed && !this.state.isDone && !this.state.hasVoted && <div> "You're Confirmed for Zach!" {this.state.error}</div>
                        }
                    </div>
                }
            </div>
        </div> 
      );
    }
  }

  export default withRouter(Vote);