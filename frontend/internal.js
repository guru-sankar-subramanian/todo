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

            // let statusHtml = ''

            row.innerHTML = `
                <td >${task.id}</td>
                <td><strong>${task.taskName}</strong></td>
                <td>${labelsHtml}</td>
                <td><span class="badge" data-variant="${statusVariant}">${task.status}</span></td>
                <td>${task.createdAt}</td>
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
    // console.log("addTask called with task name: {}", taskName);
    // console.log("addTask called with labels {}", tagInputs);

    try {
        const url = "/api/tasks";
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "taskName": taskName,
                "labels": tagInputs
            })
        });

        const data = await response.json();
        console.log();
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
    fetchTasksDataOnLoad();
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