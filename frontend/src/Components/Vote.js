import React from 'react';
import Button from 'react-bootstrap/Button'
import '../App.css';
import {PostData} from '../serviceCalls.js';

import Minush from '../Assets/Minushka.jpg'
import Zach from '../Assets/Zach.jpg'

 
class Vote extends React.Component {
    constructor(props){
      super(props);
      this.state = {
        ssn: "111110",
        dob: "12/10/1991",
        candidateClicked: -1,
        isConfirmed: false
      };

      this.whichCandidateDidIVoteFor = this.whichCandidateDidIVoteFor.bind(this);
      this.submittingMyVote = this.submittingMyVote.bind(this);
  }
  
    //hardcoded portion to return candidate string
    whichCandidateDidIVoteFor() {
      if (this.state.candidateClicked == 0 && this.state.isConfirmed == true){
        return "Minushka";
      }
      if (this.state.candidateClicked == 1 && this.state.isConfirmed == true){
        return "Zach";
      }
      else return ""; 
    }

    //Changes the candidate voted for when clicked
    minushClick = () => {
      console.log('Click!!!!');
      this.setState({
        candidateClicked : 0,
      })
    }    
    
    //Changes the candidate voted for when clicked
    zachClick = () => {
        console.log('Click!!!!');
        this.setState({
          candidateClicked : 1,
        })
    } 

    //Sends the REST call to submit vote
    submittingMyVote() {
      console.log('Submitting!'); 
      let path = "vote"; 
      let myData = {"SSN": this.state.ssn, "DOB": this.state.dob, "Candidate": this.whichCandidateDidIVoteFor()};

      PostData(path,myData).then((result) => {
        let responseJson = result;
        if(responseJson.Success == true){
            console.log (responseJson);
            console.log("--------We have voted!-----------"); 
            return true; 
        }
        else { 
            console.log (responseJson.Success);
            console.log (responseJson); 
            this.setState({ error: responseJson.Message});
            return false; }
      });
    }


    confirmSubmission = () => {
        console.log('Click!!!!');
        this.setState({
          isConfirmed: true
        })
        this.submittingMyVote(); 
    } 
  
    render () {
      return (
        <div>
            <div>
                <img src={Minush} onClick={this.minushClick} width = "300"/>
                {
                    (this.state.candidateClicked ==0) &&
                    <div>You clicked Minushka!
                        <Button onClick={this.confirmSubmission}> Confirm </Button> {
                            this.state.isConfirmed && <div> "You're Confirmed for Minushka!" </div>
                        }
                    </div>
                }
            </div>
            <div>
                <img src={Zach} onClick={this.zachClick} width = "300"/>
                {
                    (this.state.candidateClicked ==1) &&
                    <div>You clicked Zach!
                        <Button onClick={this.confirmSubmission}> Confirm </Button> {
                            this.state.isConfirmed && <div> "You're Confirmed for Zach!" </div>
                        }
                    </div>
                }
            </div>
        </div> 
      );
    }
  }

  export default Vote;