const loaderSkeletonDiv = document.getElementById('loaderSkeleton');
const taskstable = document.getElementById('tasksTable');
const taskstableBody = document.querySelector('table #tableBody');
const delay = (ms) => new Promise(timeInMs => setTimeout(timeInMs, ms));


document.addEventListener("DOMContentLoaded", function () {
    fetchTasksDataOnLoad();
});

function ariaToggle(toggle) {
    taskstable.setAttribute("aria-busy", toggle);
    taskstable.setAttribute("data-spinner", "large overlay");
    taskstable.setAttribute("aria-live", "polite");
}

// load all tasks data from the backend and populate the table
async function fetchTasksDataOnLoad() {
    ariaToggle("true");
    taskstableBody.innerHTML = "";
    const url = "/api/tasks";
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`);
        }
        const result = await response.json();
        await delay(1000)

        for (const task of result) {
            console.log(task);
            const row = document.createElement('tr');
            let statusVariant = "warning"; // Default variant
            if (task.status === "TODO") {
                statusVariant = "danger";
            } else if (task.status === "DONE") {
                statusVariant = "success";
            }


            let labelsHtml = '<span></span>';
            if (task.labels && Array.isArray(task.labels) && task.labels.length > 0) {
                labelsHtml = task.labels.map(label => `<span class="badge" data-variant="secondary">${label}</span>`).join("")
            }

            row.id = task.id;
            row.innerHTML = `
                <td>${task.id}</td>
                <td class="data-editable"><strong>${task.taskName}</strong></td>
                <td id="labels${task.id}" class="labels data-editable">${labelsHtml}</td>
                <td class="data-editable"><span class="badge" data-variant="${statusVariant}">${task.status}</span></td>
                <td>${task.createdAt}</td>
                <td>
                <button id="buttonId" class="small" data-variant="outline" commandfor="edit-dialog-form" command="show-modal" onclick="openEditTask(event)" 
                    data-id=${task.id} 
                    data-task-name="${task.taskName}"
                    data-labels=${task.labels}
                    data-status=${task.status}>Edit</button>
                </td>
            `;
            taskstableBody.appendChild(row);
        }
    } catch (error) {
        console.error(error.message);
    } finally {
        ariaToggle("false");
        taskstable.style.display = 'block';
    }
}


// handle addtask form submission
async function addTask(returnValue) {
    if (returnValue !== "save") {
        console.log("Dialog closed without saving.");
        return;
    }
    var taskName = document.querySelector("#demo-dialog-form input[name='taskName']").value;
    var tagInputs = document.getElementById("labels").value;
    callCreateOrUpdateApi(null, taskName, tagInputs, null, 'POST')
    fetchTasksDataOnLoad();
}

function openEditTask(event) {
    const rowData = event.target.dataset;
    console.log("rowData >>>>> " + JSON.stringify(rowData));
    document.querySelector("#edit-dialog-form input[name='id']").value = rowData.id;
    document.querySelector("#edit-dialog-form input[name='taskName']").value = rowData.taskName;
    document.querySelector("#edit-dialog-form ot-taginput").value = rowData.labels.split(",");
    document.querySelector("#edit-dialog-form input[name='status']").value = rowData.status;
}

function showBanner(result) {
    alertBanner = document.getElementById('alertBanner');
    alertBanner.style.display = 'block';
    if (result == 'failed') {
        alertBanner.innerHTML = " <strong>Error!</strong> Something went wrong.";
        alertBanner.setAttribute('data-variant', 'error');
    } else {
        alertBanner.innerHTML = "<strong>Task Saved Successfully!</strong>";
        alertBanner.setAttribute('data-variant', 'success');
    }

    setTimeout(() => {
        alertBanner.style.display = 'none';
    }, 1000);

}


function editTask(event) {
    var id = document.querySelector("#edit-dialog-form input[name='id']").value;
    var taskName = document.querySelector("#edit-dialog-form input[name='taskName']").value;
    var tagInputs = document.querySelector("#edit-dialog-form ot-taginput").value
    var status = document.querySelector("#edit-dialog-form select").value
    callCreateOrUpdateApi(id, taskName, tagInputs, status, 'PUT');
    fetchTasksDataOnLoad();
}

async function callCreateOrUpdateApi(id, taskName, tagInputs, statusValue, method) {
    var reqBody = "";
    if (method == 'POST') {
        reqBody = JSON.stringify({
            "taskName": taskName,
            "labels": tagInputs
        })
    } else {
        reqBody = JSON.stringify({
            "id": Number(id),
            "taskName": taskName,
            "labels": tagInputs,
            "status": statusValue
        })
    }

    try {
        const url = "/api/tasks";
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: reqBody
        });

        const data = await response.json();
        await delay(500)
        console.log(method, " call triggered :", reqBody);
        if (response.status == 201) {
            console.log('Task saved successfully:', data);
            showBanner("success")
        } else {
            console.log("Task save unsuccessful")
            showBanner("failed")
        }

    } catch (error) {
        console.error('Error saving data:', error);
    }
}