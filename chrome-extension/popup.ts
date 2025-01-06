async function getActiveTabURL() {
    const tabs = await chrome.tabs.query({
        currentWindow: true,
        active: true
    });

    return tabs[0];
}

// sendMessage example
// const onPlay = async (e) => {
//     const bookmarkTime =
//         e.target.parentNode.parentNode.getAttribute("timestamp");
//     const activeTab = await getActiveTabURL();

//     chrome.tabs.sendMessage(activeTab.id, {
//         type: "PLAY",
//         value: bookmarkTime
//     });
// };
// * HERE
const refresh = document.getElementById("refresh");
refresh.addEventListener("click", async () => {
    console.log("Refreshing");
    alert("Refresh");
    options;

    // chrome.tabs.sendMessage((tabId, tab));
});

document.addEventListener("DOMContentLoaded", async () => {
    console.log("Dom loaded");
    const activeTab = await getActiveTabURL();

    if (activeTab.url.includes("magnetdl.com")) {
        const refresh = document.getElementById("refresh");
        refresh.hidden = false;
    } else {
        const container = document.getElementById("main-text");

        container.innerHTML =
            '<div id="main-text" class="font-medium">This is not a MagnetDL page.</div>';
    }
});
