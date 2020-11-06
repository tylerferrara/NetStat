import './App.css';

export async function PostData(path, userData) { //String path, userDate in JSON form based on request 
    console.log("HERE IS MY USER DATA"); 
    console.log(userData); 
    let url = 'http://localhost:8080/'
    let res = 
        await fetch(url+path, {
            method: 'POST',
            body: JSON.stringify(userData)
       }).then((response) => response.json())
    
    return res; 
}





   
