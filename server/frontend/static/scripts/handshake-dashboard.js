$(document).ready(function() {
    function setDarkMode(isDark) {
        document.documentElement.classList.toggle("dark-mode", isDark);
        document.body.classList.toggle("dark-mode", isDark);
        localStorage.setItem("darkMode", isDark);

        // Update theme toggle button icon
        const $themeToggle = $("#theme-toggle i");
        if (isDark) {
            $themeToggle.removeClass("fa-moon").addClass("fa-sun");
        } else {
            $themeToggle.removeClass("fa-sun").addClass("fa-moon");
        }
    }

    function isDarkMode() {
        return localStorage.getItem("darkMode") === "true";
    }

    function applyDarkModeWithoutTransition(isDark) {
        document.documentElement.style.transition = "none";
        document.body.style.transition = "none";
        setDarkMode(isDark);
        document.documentElement.offsetHeight; // Force reflow
        document.documentElement.style.transition = "";
        document.body.style.transition = "";
    }

    // Apply dark mode on page load if it was previously set
    applyDarkModeWithoutTransition(isDarkMode());
    $("#page-content-wrapper").addClass("loaded");

    // Set initial theme toggle button icon
    const $themeToggleIcon = $("#theme-toggle i");
    $themeToggleIcon.toggleClass("fa-sun", isDarkMode());
    $themeToggleIcon.toggleClass("fa-moon", !isDarkMode());

    // Toggle sidebar
    $(document).on("click", "#menu-toggle", function(e) {
        e.preventDefault();
        $("#wrapper").toggleClass("toggled");
        $("body").toggleClass("sidebar-open");
    });

    // Toggle dark mode
    $(document).on("click", "#theme-toggle", function() {
        const newDarkMode = !isDarkMode();
        setDarkMode(newDarkMode);
    });

    // Attack mode options
    const attackModes = [
        { value: 0, label: "Straight" },
        { value: 1, label: "Combination" },
        { value: 3, label: "Brute-force" },
        { value: 6, label: "Hybrid Wordlist + Mask" },
        { value: 7, label: "Hybrid Mask + Wordlist" }
    ];

    // Hash mode options
    const hashModes = [
        { value: 900, label: "MD4 | Raw Hash" },
        { value: 0, label: "MD5 | Raw Hash" },
        // ... add other hash modes here if needed
    ];

    // Populate attack mode select
    attackModes.forEach(mode => {
        $("#attackMode").append(`<option value="${mode.value}">${mode.label}</option>`);
    });

    // Populate hash mode select
    hashModes.forEach(mode => {
        $("#hashMode").append(`<option value="${mode.value}">${mode.label}</option>`);
    });

    // Enable/disable wordlist based on selected attack mode
    $("#attackMode").change(function() {
        const selectedMode = parseInt($(this).val());
        // Only enable wordlist for attack modes: 0, 1, 6, 7
        $("#wordlist").prop("disabled", ![0, 1, 6, 7].includes(selectedMode));
    });

    // -------------------------------------------------------
    // Populate #clientUUID from the static variable "clientUUIDs"
    // -------------------------------------------------------
    if (typeof clientUUIDs === "string") {
        const $clientSelect = $("#clientUUID");
        // Clear any existing options
        $clientSelect.empty();

        // Optional "default" placeholder
        $clientSelect.append('<option value="">-- Select a client --</option>');

        // Split by ";"
        const uuidArray = clientUUIDs.split(";");
        uuidArray.forEach(uuid => {
            // Trim to remove any extra spaces
            const trimmed = uuid.trim();
            if (trimmed) {
                $clientSelect.append(`<option value="${trimmed}">${trimmed}</option>`);
            }
        });
    }

    // Handle "Crack" button click
    $(document).on("click", ".crack-btn", function() {
        const uuid = $(this).data("uuid");
        $("#crackUUID").val(uuid);
        $("#crackModal").modal("show");
    });

    // Handle delete button
    $(document).on("click", ".delete-btn", function() {
        const uuid = $(this).data("uuid");
        $("#deleteConfirmModal").data("uuid", uuid).modal("show");
    });

    // Confirm deletion
    $("#confirmDelete").click(function() {
        const uuid = $("#deleteConfirmModal").data("uuid");
        $.post("/delete/handshake", { uuid: uuid }, function(response) {
            if (response.success) {
                alert("Handshake deleted successfully");
                location.reload();
            } else {
                alert("Error deleting handshake: " + response.error);
            }
        });
        $("#deleteConfirmModal").modal("hide");
    });

    // Submit crack
    $("#submitCrack").click(function() {
        const formData = $("#crackForm").serialize();
        $.post("/crack/handshake", formData, function(response) {
            if (response.success) {
                alert("Crack job submitted successfully");
                $("#crackModal").modal("hide");
                location.reload();
            } else {
                alert("Error submitting crack job: " + response.error);
            }
        });
    });

    // Show hashcat options
    $(document).on("click", ".hashcat-options-btn", function() {
        const options = $(this).data("options");
        $("#hashcatOptionsContent").text(options || "No scan run");
        $("#hashcatOptionsModal").modal("show");
    });

    // Show hashcat logs
    $(document).on("click", ".hashcat-logs-btn", function() {
        const logs = $(this).data("logs");
        $("#hashcatLogsContent").text(logs || "No scan run");
        $("#hashcatLogsModal").modal("show");
    });

    // Search functionality
    $("#searchInput").on("keyup", function() {
        const searchTerm = $(this).val().toLowerCase();
        $("#handshakeTableBody tr").each(function() {
            const $row = $(this);
            const text = $row.text().toLowerCase();
            $row.toggle(text.indexOf(searchTerm) > -1);
        });
    });

    // Initialize any tooltips
    $('[data-toggle="tooltip"]').tooltip();

    // Enable wordlist by default for Straight mode (0)
    $("#wordlist").prop("disabled", false);

    // Preserve dark mode when navigating via pagination
    $(document).on("click", ".pagination .page-link", function(e) {
        e.preventDefault();
        const href = $(this).attr("href");
        if (href) {
            $("#page-content-wrapper").removeClass("loaded");
            const currentDarkMode = isDarkMode();
            $.ajax({
                url: href,
                success: function(data) {
                    const $newContent = $(data).find("#page-content-wrapper").html();
                    $("#page-content-wrapper").html($newContent);

                    // Re-apply dark mode
                    setDarkMode(currentDarkMode);
                    // Re-add loaded class with slight delay
                    setTimeout(() => {
                        $("#page-content-wrapper").addClass("loaded");
                    }, 50);
                }
            });
        }
    });

    // Check for darkMode param in URL
    const urlParams = new URLSearchParams(window.location.search);
    const darkModeParam = urlParams.get("darkMode");
    if (darkModeParam !== null) {
        setDarkMode(darkModeParam === "true");
    }
});
