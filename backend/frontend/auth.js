let isLogin = true;

function toggleMode(){

    isLogin = !isLogin;

    document.getElementById("title").innerText =
        isLogin ? "Login" : "Register";

    document.getElementById("authBtn").innerText =
        isLogin ? "Login" : "Register";

    document.querySelector(".toggle").innerText =
        isLogin
        ? "Don't have an account? Register"
        : "Already have an account? Login";
}

async function handleAuth(){

    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    const endpoint = isLogin ? "login" : "register";

    const res = await fetch(`/${endpoint}`,{
        method:"POST",
        headers:{
            "Content-Type":"application/json"
        },
        body:JSON.stringify({
            email,
            password
        })
    });

    const data = await res.json().catch(()=>null);

    if(res.ok && isLogin){

        localStorage.setItem("token",data.token);
        alert("Login successful");

        window.location.href = "index.html";
    }
    else if(res.ok){

        alert("Registered successfully. Now login.");
        toggleMode();
    }
    else{
        alert("Error occurred");
    }
}