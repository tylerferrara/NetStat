import {PostData} from "../serviceCalls"; 

export function getVotes(callback) {
    let result = {
        "Minushka": 0,
        "Zach": 0,
        "Final": false,
    };
    // get the cat's votes first
    let path = "results"; 
    PostData(path,{'Candidate': 'Minushka'}).then((res) => {
        result.Minushka = res.Votes
        result.Final = res.Final

        PostData(path,{'Candidate': 'Zach'}).then((r) => {
            result.Zach = r.Votes

            callback(result)
        });
    });
}

export async function isAllowed(ssn, dob) {

    let canLogIn = false; 
    let isRegistered = false; 

    //first check if we can log in
    let path = "login"; 
    let myData = {'SSN': ssn, 'DOB': dob};

    await PostData(path,myData).then((result) => {
        let responseJson = result;
        //if the backend says we can log in
        if(responseJson.Eligible){
            console.log (responseJson);  
            canLogIn = true; 
            isRegistered = true; 
        }
        else { 
            console.log (responseJson);   
        }
    });

    if (!canLogIn){
        let path = "register"; 
        let myData = {'SSN': ssn, 'DOB': dob};
    
        await PostData(path,myData).then((result) => {
            let responseJson = result;
            //if the backend says we can log in
            if(responseJson.Success){
                console.log (responseJson); ; 
                isRegistered = true; 
            }
            else { 
                console.log (responseJson); 
            }
        });
    }

    if (isRegistered && !canLogIn){
        let path = "login"; 
        let myData = {'SSN': ssn, 'DOB': dob};

        await PostData(path,myData).then((result) => {
            let responseJson = result;
            //if the backend says we can log in
            if(responseJson.Eligible){
                console.log (responseJson);  
                canLogIn = true; 
            }
            else { 
                console.log (responseJson);   
            }
        });

    }
    return (canLogIn && isRegistered); 
}

    





            
