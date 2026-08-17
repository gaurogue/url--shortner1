const urlInput = document.getElementById("urlInput");
const shortenBtn = document.getElementById("shortenBtn");
const loading = document.getElementById("loading");
const resultBox = document.getElementById("resultBox");
const shortUrl = document.getElementById("shortUrl");
const visitLink = document.getElementById("visitLink");
const copyBtn = document.getElementById("copyBtn");
const errorBox = document.getElementById("errorBox");
const toast = document.getElementById("toast");

// Show toast notification
function showToast(message) {
    toast.textContent = message;
    toast.classList.add("show");

    setTimeout(() => {
        toast.classList.remove("show");
    }, 2500);
}

// Validate URL
function isValidURL(url) {
    try {
        new URL(url);
        return true;
    } catch {
        return false;
    }
}

// Shorten URL
async function shortenURL() {

    const url = urlInput.value.trim();

    errorBox.classList.add("hidden");
    resultBox.classList.add("hidden");

    if (url === "") {
        errorBox.textContent = "Please enter a URL.";
        errorBox.classList.remove("hidden");
        return;
    }

    if (!isValidURL(url)) {
        errorBox.textContent = "Please enter a valid URL.";
        errorBox.classList.remove("hidden");
        return;
    }

    loading.classList.remove("hidden");
    shortenBtn.disabled = true;

    try {

       const response = await fetch("/shorten", {
            method: "POST",
            headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer " + localStorage.getItem("token")
},
            body: JSON.stringify({
                url: url
            })
        });

        if (!response.ok) {
    const errText = await response.text();
    throw new Error(errText);
}
        const data = await response.json();

        shortUrl.value = data.short_url;
        visitLink.href = data.short_url;

        resultBox.classList.remove("hidden");

    } catch (error) {

    errorBox.textContent = error.message;
    errorBox.classList.remove("hidden");

} finally {

        loading.classList.add("hidden");
        shortenBtn.disabled = false;

    }
}

// Button Click
shortenBtn.addEventListener("click", shortenURL);

// Press Enter
urlInput.addEventListener("keypress", function(event){

    if(event.key==="Enter"){
        shortenURL();
    }

});

// Copy Button
copyBtn.addEventListener("click", async () => {

    try{

        await navigator.clipboard.writeText(shortUrl.value);

        copyBtn.textContent = "Copied ✓";

        showToast("Link copied successfully!");

        setTimeout(() => {
            copyBtn.textContent = "Copy";
        },2000);

    }catch{

        showToast("Copy failed.");

    }

});