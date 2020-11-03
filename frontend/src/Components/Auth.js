import {PostData} from "../serviceCalls"; 
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

    if (isRegistered){
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

    





            
