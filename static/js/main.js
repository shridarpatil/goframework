document.addEventListener('DOMContentLoaded', function() {
    // Add field button functionality
    const addFieldBtn = document.getElementById('add-field');
    if (addFieldBtn) {
        addFieldBtn.addEventListener('click', function() {
            const fieldsDiv = document.getElementById('fields');
            const newField = document.createElement('div');
            newField.className = 'field';
            newField.innerHTML = `
                <input type="text" name="field_name" placeholder="Field Name">
                <input type="text" name="field_type" placeholder="Field Type">
                <input type="text" name="field_label" placeholder="Field Label">
                <label>
                    <input type="checkbox" name="field_required">
                    Required
                </label>
                <input type="text" name="field_permissions" placeholder="Permissions (space-separated)">
                <button type="button" class="remove-field">Remove</button>
            `;
            fieldsDiv.appendChild(newField);
        });
    }

    // Add permission button functionality
    const addPermissionBtn = document.getElementById('add-permission');
    if (addPermissionBtn) {
        addPermissionBtn.addEventListener('click', function() {
            const permissionsDiv = document.getElementById('permissions');
            const newPermission = document.createElement('div');
            newPermission.className = 'permission';
            newPermission.innerHTML = `
                <input type="text" name="permissions" placeholder="Permission">
                <button type="button" class="remove-permission">Remove</button>
            `;
            permissionsDiv.appendChild(newPermission);
        });
    }

    // Remove field and permission functionality
    document.addEventListener('click', function(e) {
        if (e.target && e.target.className == 'remove-field') {
            e.target.parentNode.remove();
        }
        if (e.target && e.target.className == 'remove-permission') {
            e.target.parentNode.remove();
        }
    });
});
