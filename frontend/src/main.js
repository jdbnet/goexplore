import { ListDir, GetConnections, SaveConnection, DeleteConnection, Delete, Rename, PromptUploadFiles, PromptUploadDirectory, PromptDownload } from '../wailsjs/go/main/App.js';

let currentConn = 'local';
let currentPath = '';
let connectionsCache = [];

function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  );
}

async function init() {
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
            username: document.getElementById('conn-username').value
        };
        const secret = document.getElementById('conn-secret').value;
        
        try {
            await SaveConnection(conn, secret);
            closeConnModal();
            loadConnections();
        } catch(err) {
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
                list.appendChild(el);
            });
        }
    } catch(e) {
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
    document.getElementById('conn-username').value = c.username || '';
    document.getElementById('conn-secret').value = ''; 
    document.getElementById('conn-delete-btn').style.display = 'block';
    
    updateProtocolFields();
    document.getElementById('conn-modal').style.display = 'flex';
};

window.openConnModal = () => {
    document.getElementById('modal-title').innerText = "Add Connection";
    document.getElementById('conn-form').reset();
    document.getElementById('conn-id').value = uuidv4();
    document.getElementById('conn-delete-btn').style.display = 'none';
    updateProtocolFields();
    document.getElementById('conn-modal').style.display = 'flex';
};

window.closeConnModal = () => {
    document.getElementById('conn-modal').style.display = 'none';
};

window.updateProtocolFields = () => {
    const protocol = document.getElementById('conn-protocol').value;
    
    const showBucket = protocol === 's3' || protocol === 'smb';
    document.getElementById('bucket-fields').style.display = showBucket ? 'block' : 'none';
    
    // Only show S3 specific fields when S3 is selected
    const showS3Specific = protocol === 's3';
    document.getElementById('region-group').style.display = showS3Specific ? 'flex' : 'none';
    document.getElementById('pathstyle-group').style.display = showS3Specific ? 'flex' : 'none';
};

window.deleteConnection = async () => {
    if(!confirm("Are you sure you want to delete this connection?")) return;
    const id = document.getElementById('conn-id').value;
    try {
        await DeleteConnection(id);
        closeConnModal();
        if(currentConn === id) {
            switchConn('local');
        } else {
            loadConnections();
        }
    } catch(err) {
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
        if (!files || files.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4">Empty directory</td></tr>';
            return;
        }
        
        files.forEach(f => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${f.is_dir ? '📁 ' : '📄 '}${f.name}</td>
                <td>${f.is_dir ? '-' : formatBytes(f.size)}</td>
                <td>${f.modified}</td>
                <td>${f.permissions}</td>
            `;
            if (f.is_dir) {
                tr.ondblclick = () => {
                    currentPath = f.path;
                    loadDirectory(currentConn, currentPath);
                };
            }
            tr.dataset.path = f.path;
            tr.dataset.name = f.name;
            tbody.appendChild(tr);
        });
    } catch(e) {
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

window.addEventListener('load', init);

// Context Menu Logic
let contextMenuTarget = null;

document.addEventListener('contextmenu', (e) => {
    const menu = document.getElementById('context-menu');
    menu.style.display = 'none';

    const row = e.target.closest('tr');
    if (row && row.dataset.path) {
        e.preventDefault();
        contextMenuTarget = {
            path: row.dataset.path,
            name: row.dataset.name
        };
        menu.style.display = 'block';
        menu.style.left = `${e.pageX}px`;
        menu.style.top = `${e.pageY}px`;
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
    if (!contextMenuTarget) return;
    
    const { path, name } = contextMenuTarget;
    
    try {
        if (action === 'delete') {
            if (confirm(`Are you sure you want to delete ${name}?`)) {
                await Delete(currentConn, path);
                refreshCurrentDir();
            }
        } else if (action === 'rename') {
            const newName = prompt(`Enter new name for ${name}:`, name);
            if (newName && newName !== name) {
                // Construct new path by replacing the last part
                const pathParts = path.split('/');
                pathParts[pathParts.length - 1] = newName;
                const newPath = pathParts.join('/');
                
                await Rename(currentConn, path, newPath);
                refreshCurrentDir();
            }
        } else if (action === 'download') {
            await PromptDownload(currentConn, path);
            showTransfers();
        }
    } catch (e) {
        alert(`Failed to ${action}: ${e}`);
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
