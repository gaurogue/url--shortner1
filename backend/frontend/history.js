const historyContainer =
    document.getElementById("historyContainer");


async function loadHistory() {

    const token = localStorage.getItem("token");

    if (!token) {

        window.location.href = "auth.html";

        return;

    }


    try {

        const response = await fetch("/my-urls", {

            method: "GET",

            headers: {

                "Authorization":
                    "Bearer " + token

            }

        });


        if (!response.ok) {

            throw new Error(
                "Failed to load URL history"
            );

        }


        const urls = await response.json();


        displayHistory(urls);


    } catch (error) {

        historyContainer.innerHTML = `
            <p>
                Unable to load your URL history.
            </p>
        `;

        console.error(error);

    }

}


function displayHistory(urls) {

    if (urls.length === 0) {

        historyContainer.innerHTML = `
            <p>
                You haven't shortened any URLs yet.
            </p>
        `;

        return;

    }


    historyContainer.innerHTML = "";


    urls.forEach(url => {

        const item =
            document.createElement("div");

        item.className = "history-item";


        item.innerHTML = `

            <div class="history-original">

                <strong>Original URL</strong>

                <p>
                    ${url.original_url}
                </p>

            </div>


            <div class="history-short">

                <strong>Short URL</strong>

                <p>
                    ${url.short_url}
                </p>

            </div>


            <div class="history-actions">

                <button
                    onclick="copyURL('${url.short_url}')"
                >
                    Copy
                </button>

              <a
    href="/redirect/${url.id}"
    target="_blank"
>
    Open
</a>

            </div>

        `;


        historyContainer.appendChild(item);

    });

}


async function copyURL(url) {

    try {

        await navigator.clipboard.writeText(url);

        alert("Short URL copied!");

    } catch {

        alert("Failed to copy URL.");

    }

}


loadHistory();