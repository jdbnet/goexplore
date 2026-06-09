import { GetVersion, ListDir, GetConnections, SaveConnection, DeleteConnection, Delete, Rename, PromptUploadFiles, PromptUploadDirectory, PromptDownload, TransferItems, GetTransfers, ClearTransfers, ReorderConnections } from '../wailsjs/go/main/App.js';

let currentConn = 'local';
let currentPath = '';
let connectionsCache = [];
let selectedItems = [];
let lastSelectedPath = null;
let showHiddenFiles = false;
let sortField = 'name';
let sortAsc = true;
let draggedConnId = null;

function uuidv4() {
    return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
        (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    );
}

async function init() {
    if (window.runtime && window.runtime.Environment) {
        const env = await window.runtime.Environment();
        if (env.platform === 'windows') {
            const nfsOpt = document.querySelector('#conn-protocol option[value="nfs"]');
            if (nfsOpt) nfsOpt.remove();
        }
    }

    try {
        const version = await GetVersion();
        const vSpan = document.getElementById('app-version');
        if (vSpan && version) {
            vSpan.innerText = version;
        }
    } catch (e) { }

    await loadConnections();
    await loadDirectory(currentConn, currentPath);

    document.getElementById('conn-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const conn = {
            id: document.getElementById('conn-id').value,
            name: document.getElementById('conn-name').value,
            protocol: document.getElementById('conn-protocol').value,
            host: document.getElementById('conn-host').value,
            port: parseInt(document.getElementById('conn-port').value) || 0,
            bucket: document.getElementById('conn-bucket').value,
            region: document.getElementById('conn-region').value,
            path_style: document.getElementById('conn-pathstyle').checked,
            secure: document.getElementById('conn-secure').checked,
            username: document.getElementById('conn-username').value
        };
        let secret = document.getElementById('conn-secret').value;
        if (conn.protocol === 'sftp' && document.getElementById('conn-sftp-auth-type').value === 'key') {
            secret = document.getElementById('conn-secret-key').value;
        }

        try {
            await SaveConnection(conn, secret);
            closeConnModal();
            loadConnections();
        } catch (err) {
            alert("Failed to save connection: " + err);
        }
    });
}

async function loadConnections() {
    const list = document.getElementById('conn-list');
    list.innerHTML = `<div class="conn-item ${currentConn === 'local' ? 'active' : ''}" onclick="switchConn('local')">
        <span class="badge">OS</span> <span style="flex:1">Local Filesystem</span>
    </div>`;

    try {
        const conns = await GetConnections();
        connectionsCache = conns || [];
        if (connectionsCache) {
            connectionsCache.forEach(c => {
                const el = document.createElement('div');
                el.className = `conn-item ${currentConn === c.id ? 'active' : ''}`;
                el.innerHTML = `<span class="badge">${c.protocol}</span> <span style="flex:1">${c.name}</span> <svg class="edit-btn" viewBox="0 0 24 24" width="16" height="16" fill="var(--accent)" style="cursor: pointer;" onclick="event.stopPropagation(); editConnection('${c.id}')"><path d="M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.06-0.94l2.03-1.58c0.18-0.14,0.23-0.41,0.12-0.61 l-1.92-3.32c-0.12-0.22-0.37-0.29-0.59-0.22l-2.39,0.96c-0.5-0.38-1.03-0.7-1.62-0.94L14.4,2.81c-0.04-0.24-0.24-0.41-0.48-0.41 h-3.84c-0.24,0-0.43,0.17-0.47,0.41L9.25,5.35C8.66,5.59,8.12,5.92,7.63,6.29L5.24,5.33c-0.22-0.08-0.47,0-0.59,0.22L2.73,8.87 C2.62,9.08,2.66,9.34,2.86,9.48l2.03,1.58C4.84,11.36,4.8,11.69,4.8,12s0.02,0.64,0.06,0.94l-2.03,1.58 c-0.18,0.14-0.23,0.41-0.12,0.61l1.92,3.32c0.12,0.22,0.37,0.29,0.59,0.22l2.39-0.96c0.5,0.38,1.03,0.7,1.62,0.94l0.36,2.54 c0.05,0.24,0.24,0.41,0.48,0.41h3.84c0.24,0,0.43-0.17,0.47-0.41l0.36-2.54c0.59-0.24,1.13-0.56,1.62-0.94l2.39,0.96 c0.22,0.08,0.47,0,0.59-0.22l1.92-3.32c0.12-0.22,0.07-0.49-0.12-0.61L19.14,12.94z M12,15.6c-1.98,0-3.6-1.62-3.6-3.6 s1.62-3.6,3.6-3.6s3.6,1.62,3.6,3.6S13.98,15.6,12,15.6z"/></svg>`;
                el.onclick = () => switchConn(c.id);
                
                el.draggable = true;
                el.addEventListener('dragstart', (e) => {
                    draggedConnId = c.id;
                    e.dataTransfer.effectAllowed = 'move';
                    el.style.opacity = '0.5';
                });
                el.addEventListener('dragend', () => {
                    draggedConnId = null;
                    el.style.opacity = '1';
                });
                el.addEventListener('dragenter', (e) => {
                    e.preventDefault();
                    if (draggedConnId && draggedConnId !== c.id) {
                        el.style.borderTop = '2px solid var(--accent)';
                    }
                });
                el.addEventListener('dragleave', (e) => {
                    el.style.borderTop = '';
                });
                el.addEventListener('dragover', (e) => {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';
                });
                el.addEventListener('drop', async (e) => {
                    e.preventDefault();
                    el.style.borderTop = '';
                    if (!draggedConnId || draggedConnId === c.id) return;
                    
                    const ids = connectionsCache.map(conn => conn.id);
                    const draggedIdx = ids.indexOf(draggedConnId);
                    const targetIdx = ids.indexOf(c.id);
                    
                    if (draggedIdx !== -1 && targetIdx !== -1) {
                        ids.splice(draggedIdx, 1);
                        ids.splice(targetIdx, 0, draggedConnId);
                        
                        try {
                            await ReorderConnections(ids);
                        } catch (err) {
                            console.error("Failed to reorder connections:", err);
                        }
                        await loadConnections();
                    }
                });

                list.appendChild(el);
            });
        }
    } catch (e) {
        console.error(e);
    }
}

window.switchConn = async (id) => {
    currentConn = id;
    currentPath = ''; // Root path
    loadConnections(); // To update 'active' class
    await loadDirectory(currentConn, currentPath);
};

window.editConnection = (id) => {
    const c = connectionsCache.find(conn => conn.id === id);
    if (!c) return;
    document.getElementById('modal-title').innerText = "Edit Connection";
    document.getElementById('conn-id').value = c.id;
    document.getElementById('conn-name').value = c.name;
    document.getElementById('conn-protocol').value = c.protocol;
    document.getElementById('conn-host').value = c.host || '';
    document.getElementById('conn-port').value = c.port || '';
    document.getElementById('conn-bucket').value = c.bucket || '';
    document.getElementById('conn-region').value = c.region || '';
    document.getElementById('conn-pathstyle').checked = c.path_style || false;
    document.getElementById('conn-secure').checked = c.secure || false;
    document.getElementById('conn-username').value = c.username || '';
    document.getElementById('conn-secret').value = '';
    document.getElementById('conn-secret-key').value = '';
    document.getElementById('conn-sftp-auth-type').value = 'password';
    document.getElementById('conn-delete-btn').style.display = 'block';

    updateProtocolFields();
    document.getElementById('conn-modal').style.display = 'flex';
};

window.openConnModal = () => {
    document.getElementById('modal-title').innerText = "Add Connection";
    document.getElementById('conn-form').reset();
    document.getElementById('conn-id').value = uuidv4();
    document.getElementById('conn-secure').checked = false;
    document.getElementById('conn-secret').value = '';
    document.getElementById('conn-secret-key').value = '';
    document.getElementById('conn-sftp-auth-type').value = 'password';
    document.getElementById('conn-delete-btn').style.display = 'none';
    updateProtocolFields();
    document.getElementById('conn-modal').style.display = 'flex';
};

window.closeConnModal = () => {
    document.getElementById('conn-modal').style.display = 'none';
};

window.updateProtocolFields = () => {
    const protocol = document.getElementById('conn-protocol').value;

    const showBucket = protocol === 's3' || protocol === 'smb' || protocol === 'nfs';
    document.getElementById('bucket-fields').style.display = showBucket ? 'block' : 'none';

    const bucketLabel = document.getElementById('bucket-label');
    if (bucketLabel) {
        if (protocol === 'nfs') {
            bucketLabel.innerText = 'Export Path';
        } else if (protocol === 'smb') {
            bucketLabel.innerText = 'Share';
        } else {
            bucketLabel.innerText = 'Bucket';
        }
    }

    const isNFS = protocol === 'nfs';
    const portGroup = document.getElementById('port-group');
    if (portGroup) portGroup.style.display = isNFS ? 'none' : 'block';
    
    const authRow = document.getElementById('auth-row');
    if (authRow) authRow.style.display = isNFS ? 'none' : 'flex';

    // Only show S3 specific fields when S3 is selected
    const showS3Specific = protocol === 's3';
    document.getElementById('region-group').style.display = showS3Specific ? 'flex' : 'none';
    document.getElementById('pathstyle-group').style.display = showS3Specific ? 'flex' : 'none';

    // FTP specific fields
    const showFTP = protocol === 'ftp';
    document.getElementById('ftp-secure-group').style.display = showFTP ? 'flex' : 'none';

    // SFTP Auth Type toggle
    const authTypeSelect = document.getElementById('conn-sftp-auth-type');
    const secretInput = document.getElementById('conn-secret');
    const secretKeyArea = document.getElementById('conn-secret-key');
    const secretLabel = document.getElementById('secret-label');

    if (protocol === 'sftp') {
        authTypeSelect.style.display = 'block';
        if (authTypeSelect.value === 'key') {
            secretInput.style.display = 'none';
            secretKeyArea.style.display = 'block';
            secretLabel.innerText = 'SSH Private Key';
        } else {
            secretInput.style.display = 'block';
            secretKeyArea.style.display = 'none';
            secretLabel.innerText = 'Password';
        }
    } else {
        authTypeSelect.style.display = 'none';
        secretInput.style.display = 'block';
        secretKeyArea.style.display = 'none';
        secretLabel.innerText = 'Password / Secret Key';
    }
};

window.deleteConnection = async () => {
    if (!confirm("Are you sure you want to delete this connection?")) return;
    const id = document.getElementById('conn-id').value;
    try {
        await DeleteConnection(id);
        closeConnModal();
        if (currentConn === id) {
            switchConn('local');
        } else {
            loadConnections();
        }
    } catch (err) {
        alert("Failed to delete connection: " + err);
    }
};

async function loadDirectory(connId, path) {
    const tbody = document.getElementById('file-list');
    tbody.innerHTML = '<tr><td colspan="4">Loading...</td></tr>';
    document.getElementById('breadcrumb').innerText = path || '/';

    try {
        const files = await ListDir(connId, path);
        tbody.innerHTML = '';
        selectedItems = [];
        lastSelectedPath = null;

        let displayFiles = files || [];
        if (!showHiddenFiles) {
            displayFiles = displayFiles.filter(f => !f.name.startsWith('.'));
        }

        if (displayFiles.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4">Empty directory</td></tr>';
            return;
        }

        displayFiles.sort((a, b) => {
            if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1;

            let cmp = 0;
            switch (sortField) {
                case 'name':
                    cmp = a.name.localeCompare(b.name);
                    break;
                case 'size':
                    cmp = a.size - b.size;
                    break;
                case 'modified':
                    cmp = new Date(a.modified).getTime() - new Date(b.modified).getTime();
                    if (isNaN(cmp)) cmp = 0;
                    break;
                case 'permissions':
                    cmp = a.permissions.localeCompare(b.permissions);
                    break;
            }
            return sortAsc ? cmp : -cmp;
        });

        displayFiles.forEach(f => {
            const tr = document.createElement('tr');
            tr.dataset.path = f.path;
            tr.dataset.name = f.name;
            tr.dataset.isDir = f.is_dir;
            tr.dataset.size = f.size;

            tr.innerHTML = `
                <td>${f.is_dir ? '📁 ' : '📄 '}${f.name}</td>
                <td>${f.is_dir ? '-' : formatBytes(f.size)}</td>
                <td>${formatDate(f.modified)}</td>
                <td>${f.permissions}</td>
            `;

            tr.onclick = (e) => {
                const path = f.path;
                const isSelected = selectedItems.some(i => i.path === path);

                if (e.ctrlKey || e.metaKey) {
                    if (isSelected) {
                        selectedItems = selectedItems.filter(i => i.path !== path);
                        tr.classList.remove('selected');
                    } else {
                        selectedItems.push(f);
                        tr.classList.add('selected');
                    }
                    lastSelectedPath = path;
                } else if (e.shiftKey && lastSelectedPath) {
                    const allRows = Array.from(tbody.querySelectorAll('tr'));
                    const lastIdx = allRows.findIndex(r => r.dataset.path === lastSelectedPath);
                    const currIdx = allRows.findIndex(r => r.dataset.path === path);
                    const startIdx = Math.min(lastIdx, currIdx);
                    const endIdx = Math.max(lastIdx, currIdx);

                    selectedItems = [];
                    allRows.forEach(r => r.classList.remove('selected'));

                    for (let i = startIdx; i <= endIdx; i++) {
                        const row = allRows[i];
                        if (row.dataset.path) {
                            row.classList.add('selected');
                            selectedItems.push({
                                path: row.dataset.path,
                                name: row.dataset.name,
                                is_dir: row.dataset.isDir === 'true',
                                size: parseInt(row.dataset.size) || 0
                            });
                        }
                    }
                } else {
                    selectedItems = [f];
                    Array.from(tbody.querySelectorAll('tr')).forEach(r => r.classList.remove('selected'));
                    tr.classList.add('selected');
                    lastSelectedPath = path;
                }
            };

            if (f.is_dir) {
                tr.ondblclick = () => {
                    currentPath = f.path;
                    loadDirectory(currentConn, currentPath);
                };
            }
            tbody.appendChild(tr);
        });
    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="4" style="color: var(--error)">Error: ${e}</td></tr>`;
    }
}

window.refreshCurrentDir = () => {
    loadDirectory(currentConn, currentPath);
};

window.showTransfers = () => {
    document.getElementById('transfer-tray').style.display = 'block';
};
window.hideTransfers = () => {
    document.getElementById('transfer-tray').style.display = 'none';
};

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    if (!dateString) return '-';
    const d = new Date(dateString);
    if (isNaN(d.getTime())) return dateString;
    return `${String(d.getDate()).padStart(2, '0')}/${String(d.getMonth() + 1).padStart(2, '0')}/${d.getFullYear()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

window.setSort = (field) => {
    if (sortField === field) {
        sortAsc = !sortAsc;
    } else {
        sortField = field;
        sortAsc = true;
    }
    ['name', 'size', 'modified', 'permissions'].forEach(f => {
        document.getElementById(`sort-ind-${f}`).innerText = '';
    });
    document.getElementById(`sort-ind-${field}`).innerText = sortAsc ? '↑' : '↓';
    refreshCurrentDir();
};

window.clearTransfers = async () => {
    try {
        await ClearTransfers();
        pollTransfers();
    } catch (e) {
        console.error(e);
    }
};

async function pollTransfers() {
    try {
        const transfers = await GetTransfers();
        const list = document.getElementById('transfer-list');
        const btn = document.getElementById('transfers-btn');
        btn.innerText = `Transfers (${transfers ? transfers.length : 0})`;

        if (!transfers || transfers.length === 0) {
            list.innerHTML = '<div style="color: var(--text-secondary);">No active transfers</div>';
            return;
        }

        let html = '';
        transfers.forEach(t => {
            const pct = t.bytes_total > 0 ? (t.bytes_done / t.bytes_total) * 100 : 0;
            const speed = t.status === 'active' ? `${t.speed_mbps.toFixed(2)} MB/s` : t.status;
            const eta = t.status === 'active' ? `${t.eta_seconds}s remaining` : '';
            const color = t.status === 'failed' ? 'var(--error)' : (t.status === 'complete' ? 'var(--success)' : 'var(--accent)');

            html += `
            <div class="transfer-item">
                <div style="width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-size: 0.85rem;" title="${t.filename}">${t.filename}</div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${pct}%; background: ${color}"></div>
                </div>
                <div class="transfer-meta">
                    <div>${formatBytes(t.bytes_done)} / ${formatBytes(t.bytes_total)}</div>
                    <div style="color: ${color}">${speed} ${eta}</div>
                    ${t.error ? `<div style="color: var(--error); font-size: 0.7rem;">${t.error}</div>` : ''}
                </div>
            </div>`;
        });
        list.innerHTML = html;
    } catch (e) {
        console.error(e);
    }
}

setInterval(pollTransfers, 1000);

window.addEventListener('load', init);

// Context Menu Logic
document.addEventListener('contextmenu', (e) => {
    const menu = document.getElementById('context-menu');
    menu.style.display = 'none';

    document.getElementById('ctx-hidden-text').innerText = showHiddenFiles ? 'Hide Hidden Files' : 'Show Hidden Files';

    const row = e.target.closest('tr');
    const isBrowserPane = e.target.closest('#browser-pane');

    if (row && row.dataset.path && !row.closest('#transfer-dest-list')) {
        e.preventDefault();

        if (!selectedItems.some(i => i.path === row.dataset.path)) {
            selectedItems = [{
                path: row.dataset.path,
                name: row.dataset.name,
                is_dir: row.dataset.isDir === 'true',
                size: parseInt(row.dataset.size) || 0
            }];
            const tbody = document.getElementById('file-list');
            Array.from(tbody.querySelectorAll('tr')).forEach(r => r.classList.remove('selected'));
            row.classList.add('selected');
            lastSelectedPath = row.dataset.path;
        }

        document.getElementById('ctx-rename').style.display = 'block';
        document.getElementById('ctx-transfer').style.display = 'block';
        document.getElementById('ctx-download').style.display = 'block';
        document.getElementById('ctx-delete').style.display = 'block';
        document.getElementById('ctx-divider').style.display = 'block';

        menu.style.display = 'block';
        
        let left = e.pageX;
        let top = e.pageY;
        
        if (left + menu.offsetWidth > window.innerWidth) {
            left = window.innerWidth - menu.offsetWidth;
        }
        if (top + menu.offsetHeight > window.innerHeight) {
            top = window.innerHeight - menu.offsetHeight;
        }
        
        menu.style.left = `${left}px`;
        menu.style.top = `${top}px`;
    } else if (isBrowserPane && !e.target.closest('#transfer-dest-list')) {
        e.preventDefault();

        document.getElementById('ctx-rename').style.display = 'none';
        document.getElementById('ctx-transfer').style.display = 'none';
        document.getElementById('ctx-download').style.display = 'none';
        document.getElementById('ctx-delete').style.display = 'none';
        document.getElementById('ctx-divider').style.display = 'none';

        menu.style.display = 'block';
        
        let left = e.pageX;
        let top = e.pageY;
        
        if (left + menu.offsetWidth > window.innerWidth) {
            left = window.innerWidth - menu.offsetWidth;
        }
        if (top + menu.offsetHeight > window.innerHeight) {
            top = window.innerHeight - menu.offsetHeight;
        }
        
        menu.style.left = `${left}px`;
        menu.style.top = `${top}px`;
    }
});

document.addEventListener('click', () => {
    document.getElementById('context-menu').style.display = 'none';
    const uploadMenu = document.getElementById('upload-menu');
    if (uploadMenu) uploadMenu.style.display = 'none';
});

window.toggleUploadMenu = (e) => {
    if (e) e.stopPropagation();
    const menu = document.getElementById('upload-menu');
    if (menu.style.display === 'none') {
        menu.style.display = 'block';
    } else {
        menu.style.display = 'none';
    }
};

window.handleContextMenu = async (action) => {
    if (action === 'toggle_hidden') {
        showHiddenFiles = !showHiddenFiles;
        refreshCurrentDir();
        return;
    }

    if (selectedItems.length === 0) return;

    try {
        if (action === 'delete') {
            if (confirm(`Are you sure you want to delete ${selectedItems.length} items?`)) {
                for (const item of selectedItems) {
                    await Delete(currentConn, item.path);
                }
                refreshCurrentDir();
            }
        } else if (action === 'rename') {
            if (selectedItems.length > 1) {
                alert("Cannot rename multiple items at once.");
                return;
            }
            const item = selectedItems[0];
            const newName = prompt(`Enter new name for ${item.name}:`, item.name);
            if (newName && newName !== item.name) {
                const pathParts = item.path.split('/');
                pathParts[pathParts.length - 1] = newName;
                const newPath = pathParts.join('/');

                await Rename(currentConn, item.path, newPath);
                refreshCurrentDir();
            }
        } else if (action === 'download') {
            for (const item of selectedItems) {
                await PromptDownload(currentConn, item.path);
            }
            showTransfers();
        } else if (action === 'transfer') {
            openTransferModal();
        }
    } catch (e) {
        alert(`Failed to ${action}: ${e}`);
    }
};

window.navigateUp = () => {
    if (!currentPath || currentPath === '/') return;
    const parts = currentPath.split('/').filter(p => p);
    parts.pop();
    currentPath = parts.length ? '/' + parts.join('/') : '';
    loadDirectory(currentConn, currentPath);
};

let transferDestConn = 'local';
let transferDestPath = '';

window.openTransferModal = () => {
    const select = document.getElementById('transfer-dest-conn');
    select.innerHTML = `<option value="local">Local Filesystem</option>`;
    connectionsCache.forEach(c => {
        select.innerHTML += `<option value="${c.id}">${c.name}</option>`;
    });
    document.getElementById('transfer-modal').style.display = 'flex';
    loadTransferDestRoot();
};

window.closeTransferModal = () => {
    document.getElementById('transfer-modal').style.display = 'none';
    document.getElementById('transfer-verify').checked = false;
    document.getElementById('transfer-limit').value = "0";
};

window.loadTransferDestRoot = () => {
    transferDestConn = document.getElementById('transfer-dest-conn').value;
    transferDestPath = '';
    loadTransferDestDirectory(transferDestConn, transferDestPath);
};

window.loadTransferDestDirectory = async (connId, path) => {
    transferDestPath = path;
    const tbody = document.getElementById('transfer-dest-list');
    tbody.innerHTML = '<tr><td colspan="4">Loading...</td></tr>';
    document.getElementById('transfer-dest-breadcrumb').innerText = path || '/';

    try {
        const files = await ListDir(connId, path);
        tbody.innerHTML = '';

        if (path && path !== '/') {
            const tr = document.createElement('tr');
            tr.innerHTML = `<td>📁 ..</td>`;
            tr.style.cursor = 'pointer';
            tr.ondblclick = () => {
                const parts = path.split('/').filter(p => p);
                parts.pop();
                loadTransferDestDirectory(connId, parts.length ? '/' + parts.join('/') : '');
            };
            tbody.appendChild(tr);
        }

        if (!files || files.length === 0) {
            tbody.innerHTML += '<tr><td>Empty directory</td></tr>';
            return;
        }

        files.forEach(f => {
            const tr = document.createElement('tr');
            tr.style.cursor = 'pointer';
            tr.innerHTML = `<td>${f.is_dir ? '📁 ' : '📄 '}${f.name}</td>`;
            if (f.is_dir) {
                tr.ondblclick = () => {
                    loadTransferDestDirectory(connId, f.path);
                };
            }
            tbody.appendChild(tr);
        });
    } catch (e) {
        tbody.innerHTML = `<tr><td style="color: var(--error)">Error: ${e}</td></tr>`;
    }
};

window.executeTransfer = async () => {
    if (selectedItems.length === 0) return;
    try {
        const verify = document.getElementById('transfer-verify').checked;
        const limit = parseInt(document.getElementById('transfer-limit').value, 10) || 0;
        await TransferItems(currentConn, transferDestConn, transferDestPath, selectedItems, verify, limit);
        closeTransferModal();
        showTransfers();
        selectedItems = [];
        Array.from(document.querySelectorAll('#file-list tr')).forEach(r => r.classList.remove('selected'));
    } catch (e) {
        alert("Transfer failed: " + e);
    }
};

window.promptUploadFiles = async () => {
    try {
        await PromptUploadFiles(currentConn, currentPath);
        showTransfers();
    } catch (e) {
        alert("Upload failed: " + e);
    }
};

window.promptUploadDirectory = async () => {
    try {
        await PromptUploadDirectory(currentConn, currentPath);
        showTransfers();
    } catch (e) {
        alert("Upload failed: " + e);
    }
};
