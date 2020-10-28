import React from 'react';
import Button from 'react-bootstrap/Button'
import '../App.css';

import Minush from '../Assets/Minushka.jpg'
import Zach from '../Assets/Zach.jpg'

 
class Vote extends React.Component {
    constructor(props){
      super(props);
      this.state = {
        minushClicked : false,
        zachClicked : false,
        isConfirmed: false
      }
    }
  
    minushClick = () => {
      console.log('Click!!!!');
      this.setState({
        minushClicked: true
      })
    }    
    
    zachClick = () => {
        console.log('Click!!!!');
        this.setState({
          zachClicked: true
        })
    } 

    confirmSubmission = () => {
        console.log('Click!!!!');
        this.setState({
          isConfirmed: true
        })
    } 
  
    render () {
      return (
        <div>
            <div>
                <img src={Minush} onClick={this.minushClick} width = "300"/>
                {
                    this.state.minushClicked &&
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
                    this.state.zachClicked &&
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