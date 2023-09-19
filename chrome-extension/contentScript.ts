(() => {
    chrome.runtime.onMessage.addListener((obj, sender, response) => {
        const { type, value, server } = obj;

        if (type === "VALID" || type === "REFRESH") {
            refreshData(server);
        }
    });

    //     debugButton.addEventListener("click", async () => {
    //         console.log("click");
    //         let [tab] = await chrome.tabs.query({
    //             active: true,
    //             currentWindow: true
    //         });
    //         chrome.scripting.executeScript({
    //             target: { tabId: tab.id },
    //             func: initStorageCache
    //         });
    //     });

    //     function initStorageCache() {
    //         console.log("Refreshing cache");
    //         chrome.storage.sync.get("options", (data) => {
    //             Object.assign(options, data.options);
    //             console.log(options.serverUrl);
    //         });
    //     }
    function refreshData(target: string) {
        console.log("Refresh data in contentScript.js");
        const magnetLinks = document.getElementsByClassName("m");
        Array.from(magnetLinks, (magnet) => {
            console.log(magnet);
            const link = magnet.querySelector("a").href;
            console.log(link);
            const button = makeButton(target, link);
            magnet.append(button);
        });
    }
    function makeButton(target, magnet) {
        const magnetData: Magnet = { link: magnet };
        const icon = document.createElement("img");
        icon.src = chrome.runtime.getURL("/images/cloud-storage-16x16.png");
        icon.onclick = () => {
            icon.src = chrome.runtime.getURL("/images/loading.png");
            postData(target, magnetData).then((response) => {
                if (response.status == 200) {
                    icon.src = chrome.runtime.getURL(
                        "/images/cloud-storage-16x16.png"
                    );
                } else {
                    icon.src = chrome.runtime.getURL("/images/error.png");
                    console.log(response);
                }
            });
        };
        return icon;
    }
    async function postData(target: string, data: Object) {
        const response = await fetch(`${target}/api/magnet`, {
            method: "POST",
            body: JSON.stringify(data)
        });
        return response.json();
    }
})();
