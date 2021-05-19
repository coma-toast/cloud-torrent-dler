let downloadButtons = document.getElementsByClassName("download")
// for (let button of downloadButtons) {
//     console.log("hi")
//     console.log(button)
// }
Array.from(downloadButtons).forEach((button) => {
    button.addEventListener("click", function() {
        let formData = new FormData
        formData.append("link", this.dataset.torrent)
        let data = JSON.stringify(Object.fromEntries(formData))
        button.classList.remove("btn-success")
        button.classList.add("btn-danger")
        let resolution = button.innerHTML
        button.innerHTML = `<span class="spinner-border spinner-border-sm"></span>`
        apiRequest("POST", "torrent", data)
        .then((response) => response.json())
        .then(responseData => {
            if (responseData.result) {
                button.innerHTML = resolution
                button.classList.remove("btn-danger")
                button.classList.add("btn-warning")
            }
            console.log(responseData)
            })
    })
});


/**
 * * This is the main API call function 
 *
 * @param   {string}                method  "GET", "POST"
 * @param   {string}                target
 * @param   {FormData}              data  new FormData
 * @param   {Object|string[[]]}     params   {clientID: 1} | [['clientID', '1']]
 *
 * @return  {Response}
 */
async function apiRequest(method, target, data) {
    try {
        let response = new Response();
        let headers = new Headers();
        /*
         * Fetch allows relative URLs, but you can't have a body for GET requests.
         * So we have to build full URL with query params for GET requests with no body (not even `null`)
         * and then use the normal fetch request for POSTs
         */
        if (method === "GET") {
            params = new URLSearchParams(data).toString();
            url = new URL(
                window.location.origin +
                    "/api/" +
                    target +
                    "?" +
                    params
            );
            response = await fetch(url, {
                method: method,
                mode: "cors",
                headers: {
                    headers,
                },
            });
        } else {
            response = await fetch("/api/" + target, {
                method: method,
                mode: "cors",
                headers: {
                    headers,
                },
                body: data,
            });
        }
        console.debug(response);

        return response;
    } catch (error) {
        console.error(error);
    }
}
