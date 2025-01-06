type Options = {
    serverUrl: string;
};

type Magnet = {
    link: string;
};

// Where we will expose all the data we retrieve from storage.sync.
const options: Options = {
    serverUrl: ""
};

chrome.tabs.onUpdated.addListener((tabId, tab) => {
    initStorageCache();
    if (tab.url && tab.url.includes("magnetdl.com")) {
        chrome.tabs.sendMessage(tabId, {
            type: "VALID",
            server: options.serverUrl
        });
    }
});

function initStorageCache() {
    console.log("Refreshing cache");
    chrome.storage.sync.get("options", (data) => {
        Object.assign(options, data.options);
        console.log(options.serverUrl);
    });
}

// // Asynchronously retrieve data from storage.sync, then cache it.
// const initStorageCache = getAllStorageSyncData()
//     .then((items) => {
//         // Copy the data retrieved from storage into storageCache.
//         Object.assign(options, items);
//     })
//     .then(() => {
//         console.log("Data:", options);
//     });

chrome.action.onClicked.addListener(async (tab) => {
    try {
        await initStorageCache;
    } catch (e) {
        // Handle error that occurred during storage initialization.
    }
    // Normal action handler logic.
    let target = document.getElementById("target") as HTMLInputElement;
    console.log(target);
    target.value = options.serverUrl;
});

// Reads all data out of storage.sync and exposes it via a promise.
//
// Note: Once the Storage API gains promise support, this function
// can be greatly simplified.
function getAllStorageSyncData() {
    // Immediately return a promise and start asynchronous work
    return new Promise((resolve, reject) => {
        // Asynchronously fetch all data from storage.sync.
        chrome.storage.sync.get(null, (items) => {
            // Pass any observed errors down the promise chain.
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            // Pass the data retrieved from storage down the promise chain.
            resolve(items);
        });
    });
}
