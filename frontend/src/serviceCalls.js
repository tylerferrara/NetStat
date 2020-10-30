import './App.css';

export function PostData(path, userData) { //String path, userDate in JSON form based on request 
    let url = 'http://localhost:8080/'
    return new Promise((resolve, reject) =>{
        fetch(url+path, {
            method: 'POST',
            body: JSON.stringify(userData)
       }).then((response) => response.json()).then((res) => {resolve(res); })
       .catch((error) => { reject(error); });
    })
}





   
