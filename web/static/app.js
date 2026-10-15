// State Management
let currentUser = null;
let currentResidentFilter = 'All';
let allAdminTickets = [];
let categoryChartInstance = null;

// Initialize on page load
document.addEventListener('DOMContentLoaded', async () => {
    await checkCurrentUser();
});

// Check logged in status
async function checkCurrentUser() {
    try {
        const res = await fetch('/api/v1/auth/me');
        if (res.ok) {
            const data = await res.json();
            currentUser = data.data;
            renderUserInterface();
        } else {
            currentUser = null;
            renderGuestInterface();
        }
    } catch (err) {
        currentUser = null;
        renderGuestInterface();
    }
}

function renderUserInterface() {
    document.getElementById('guestHero').classList.add('hidden');
    document.getElementById('guestNav').classList.add('hidden');
    document.getElementById('userProfileNav').classList.remove('hidden');

    document.getElementById('navUserName').innerText = currentUser.full_name;
    const roleBadge = document.getElementById('navUserRoleBadge');

    if (currentUser.role === 'admin' || currentUser.role === 'technician') {
        roleBadge.className = 'text-xs px-2 py-0.5 rounded-full inline-block font-semibold bg-amber-100 text-amber-800';
        roleBadge.innerText = currentUser.role === 'admin' ? '🛡️ นิติบุคคล (Admin)' : '🔧 ช่างซ่อมบำรุง';
        document.getElementById('residentPortal').classList.add('hidden');
        document.getElementById('adminPortal').classList.remove('hidden');
        loadAdminDashboard();
        loadTechniciansList();
    } else {
        roleBadge.className = 'text-xs px-2 py-0.5 rounded-full inline-block font-semibold bg-emerald-100 text-emerald-800';
        roleBadge.innerText = `🚪 ลูกบ้าน (ห้อง ${currentUser.room_no})`;
        document.getElementById('residentHeaderName').innerText = currentUser.full_name;
        document.getElementById('residentHeaderRoom').innerText = currentUser.room_no;
        document.getElementById('ctRoomNo').value = currentUser.room_no;
        document.getElementById('adminPortal').classList.add('hidden');
        document.getElementById('residentPortal').classList.remove('hidden');
        loadResidentTickets();
    }
}

function renderGuestInterface() {
    document.getElementById('guestHero').classList.remove('hidden');
    document.getElementById('guestNav').classList.remove('hidden');
    document.getElementById('userProfileNav').classList.add('hidden');
    document.getElementById('residentPortal').classList.add('hidden');
    document.getElementById('adminPortal').classList.add('hidden');
}

// Quick Demo Login
async function demoLogin(role) {
    let identifier = 'resident@condo.local';
    let password = 'password123';

    if (role === 'admin') {
        identifier = 'admin@condo.local';
        password = 'admin123';
    }

    try {
        const res = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ identifier, password })
        });
        const data = await res.json();
        if (data.success) {
            currentUser = data.data.user;
            closeModal('authModal');
            renderUserInterface();
            Swal.fire({
                icon: 'success',
                title: 'เข้าสู่ระบบสำเร็จ',
                text: `ยินดีต้อนรับ ${currentUser.full_name}`,
                timer: 1500,
                showConfirmButton: false
            });
        } else {
            Swal.fire('ข้อผิดพลาด', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้', 'error');
    }
}

// Standard Login & Register
async function handleLogin(e) {
    e.preventDefault();
    const identifier = document.getElementById('loginIdentifier').value;
    const password = document.getElementById('loginPassword').value;

    try {
        const res = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ identifier, password })
        });
        const data = await res.json();
        if (data.success) {
            currentUser = data.data.user;
            closeModal('authModal');
            renderUserInterface();
            Swal.fire({ icon: 'success', title: 'เข้าสู่ระบบสำเร็จ', timer: 1500, showConfirmButton: false });
        } else {
            Swal.fire('เข้าสู่ระบบไม่สำเร็จ', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const req = {
        full_name: document.getElementById('regFullName').value,
        room_no: document.getElementById('regRoomNo').value,
        phone_number: document.getElementById('regPhone').value,
        email: document.getElementById('regEmail').value,
        password: document.getElementById('regPassword').value
    };

    try {
        const res = await fetch('/api/v1/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(req)
        });
        const data = await res.json();
        if (data.success) {
            currentUser = data.data.user;
            closeModal('authModal');
            renderUserInterface();
            Swal.fire({ icon: 'success', title: 'ลงทะเบียนสำเร็จ', text: 'ยินดีต้อนรับสู่ CondoCare!', timer: 1500, showConfirmButton: false });
        } else {
            Swal.fire('ลงทะเบียนไม่สำเร็จ', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    }
}

async function logout() {
    await fetch('/api/v1/auth/logout', { method: 'POST' });
    currentUser = null;
    renderGuestInterface();
    Swal.fire({ icon: 'info', title: 'ออกจากระบบแล้ว', timer: 1200, showConfirmButton: false });
}

// ================= RESIDENT MODULE =================
function switchResidentTab(tab) {
    currentResidentFilter = tab;
    ['All', 'Pending', 'In_Progress', 'Completed', 'Closed'].forEach(t => {
        const el = document.getElementById(`resTab${t}`);
        if (el) {
            if (t === tab) {
                el.className = 'tab-active py-2 px-1 focus:outline-none transition';
            } else {
                el.className = 'text-slate-500 hover:text-slate-700 py-2 px-1 focus:outline-none transition';
            }
        }
    });
    loadResidentTickets();
}

async function loadResidentTickets() {
    const listEl = document.getElementById('residentTicketsList');
    listEl.innerHTML = `<div class="text-center py-8 text-slate-400"><i class="fa-solid fa-spinner fa-spin text-2xl mb-2 text-brand-600"></i><p>กำลังโหลดข้อมูล...</p></div>`;

    try {
        const url = `/api/v1/tickets?status=${currentResidentFilter === 'All' ? '' : currentResidentFilter}`;
        const res = await fetch(url);
        const data = await res.json();

        if (!data.success || !data.data || data.data.length === 0) {
            listEl.innerHTML = `
                <div class="text-center py-12 text-slate-400">
                    <i class="fa-solid fa-folder-open text-4xl mb-3 text-slate-300"></i>
                    <p class="font-medium text-slate-600">ไม่พบรายการแจ้งซ่อมในหมวดนี้</p>
                    <p class="text-xs text-slate-400 mt-1">กดปุ่ม "แจ้งซ่อมปัญหาใหม่" เพื่อเริ่มต้นส่งเรื่อง</p>
                </div>`;
            return;
        }

        let html = '';
        data.data.forEach(t => {
            const statusBadge = getStatusBadge(t.status);
            const urgencyBadge = getUrgencyBadge(t.urgency);
            const categoryName = getCategoryName(t.category);
            const dateStr = new Date(t.created_at).toLocaleDateString('th-TH', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });

            // Image gallery
            let imagesHtml = '';
            if (t.images && t.images.length > 0) {
                imagesHtml = `<div class="flex flex-wrap gap-2 mt-3">`;
                t.images.forEach(img => {
                    imagesHtml += `<a href="${img.file_url}" target="_blank" class="block w-16 h-16 rounded-lg overflow-hidden border border-slate-200 shadow-sm hover:opacity-80 transition"><img src="${img.file_url}" class="w-full h-full object-cover"></a>`;
                });
                imagesHtml += `</div>`;
            }

            // Tech remark
            let techNoteHtml = '';
            if (t.internal_note) {
                techNoteHtml = `
                    <div class="mt-3 p-3 bg-amber-50 rounded-xl border border-amber-200 text-xs text-amber-900">
                        <span class="font-bold flex items-center gap-1 mb-1 text-amber-800">
                            <i class="fa-solid fa-wrench"></i> บันทึกจากช่างผู้รับผิดชอบ:
                        </span>
                        ${escapeHtml(t.internal_note)}
                    </div>`;
            }

            // Action Buttons based on Status
            let actionsHtml = '';
            if (t.status === 'Pending') {
                actionsHtml = `
                    <div class="flex items-center gap-2 mt-4 pt-3 border-t border-slate-100">
                        <button onclick="openEditModal('${t.id}', '${escapeAttr(t.title)}', '${t.category}', '${t.urgency}', '${escapeAttr(t.description)}')" class="px-3 py-1.5 text-xs font-semibold bg-amber-50 text-amber-700 hover:bg-amber-100 rounded-lg transition flex items-center gap-1">
                            <i class="fa-solid fa-pen-to-square"></i> แก้ไขใบแจ้งซ่อม
                        </button>
                        <button onclick="confirmCancelTicket('${t.id}')" class="px-3 py-1.5 text-xs font-semibold bg-rose-50 text-rose-700 hover:bg-rose-100 rounded-lg transition flex items-center gap-1">
                            <i class="fa-solid fa-ban"></i> ยกเลิกใบแจ้งซ่อม
                        </button>
                    </div>`;
            } else if (t.status === 'Completed' && !t.feedback) {
                // Form 2 trigger: Completed status only!
                actionsHtml = `
                    <div class="flex items-center gap-2 mt-4 pt-3 border-t border-slate-100">
                        <button onclick="openFeedbackModal('${t.id}')" class="px-4 py-2 text-xs font-semibold bg-gradient-to-r from-teal-500 to-emerald-600 hover:from-teal-600 hover:to-emerald-700 text-white rounded-xl shadow-md transition flex items-center gap-1.5 animate-pulse">
                            <i class="fa-solid fa-star text-amber-300"></i> ให้คะแนนและรีวิวงานซ่อม (Form 2)
                        </button>
                    </div>`;
            } else if (t.feedback) {
                actionsHtml = `
                    <div class="mt-3 p-3 bg-emerald-50 rounded-xl border border-emerald-200 text-xs text-emerald-900">
                        <div class="flex items-center justify-between font-bold mb-1">
                            <span class="flex items-center gap-1 text-emerald-800"><i class="fa-solid fa-circle-check"></i> คุณได้ประเมินงานซ่อมแล้ว:</span>
                            <span class="text-amber-500">${'★'.repeat(t.feedback.rating)}${'☆'.repeat(5 - t.feedback.rating)}</span>
                        </div>
                        <p class="italic text-emerald-700">"${escapeHtml(t.feedback.comment || 'ไม่มีข้อความเพิ่มเติม')}"</p>
                    </div>`;
            }

            html += `
                <div class="py-5">
                    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2">
                        <div class="flex items-center gap-2">
                            <span class="font-mono text-xs font-bold px-2 py-0.5 rounded bg-slate-100 text-slate-700">${t.ticket_no}</span>
                            <span class="text-xs text-slate-400">• ${dateStr}</span>
                            ${urgencyBadge}
                        </div>
                        <div>${statusBadge}</div>
                    </div>
                    <h4 class="text-base font-bold text-slate-900 mt-1">${escapeHtml(t.title)}</h4>
                    <p class="text-xs font-semibold text-brand-700 mt-0.5"><i class="fa-solid fa-tag mr-1"></i>หมวดหมู่: ${categoryName}</p>
                    <p class="text-xs text-slate-600 mt-2 leading-relaxed bg-slate-50/70 p-3 rounded-xl border border-slate-100">${escapeHtml(t.description)}</p>
                    ${imagesHtml}
                    ${techNoteHtml}
                    ${actionsHtml}
                </div>`;
        });

        listEl.innerHTML = html;
    } catch (err) {
        listEl.innerHTML = `<div class="text-center py-8 text-rose-500">ไม่สามารถโหลดรายการได้ กรุณาลองใหม่อีกครั้ง</div>`;
    }
}

// FORM 1: Create Ticket Action
async function handleCreateTicket(e) {
    e.preventDefault();
    const btn = document.getElementById('btnSubmitTicket');
    btn.disabled = true;
    btn.innerHTML = `<i class="fa-solid fa-spinner fa-spin mr-1"></i> กำลังส่งข้อมูล...`;

    const formData = new FormData();
    formData.append('title', document.getElementById('ctTitle').value);
    formData.append('category', document.getElementById('ctCategory').value);
    formData.append('urgency', document.getElementById('ctUrgency').value);
    formData.append('description', document.getElementById('ctDescription').value);

    const fileInput = document.getElementById('ctImages');
    if (fileInput.files.length > 0) {
        for (let i = 0; i < fileInput.files.length; i++) {
            formData.append('images', fileInput.files[i]);
        }
    }

    try {
        const res = await fetch('/api/v1/tickets', {
            method: 'POST',
            body: formData
        });
        const data = await res.json();
        if (data.success) {
            closeModal('createTicketModal');
            document.getElementById('createTicketForm').reset();
            loadResidentTickets();
            Swal.fire({
                icon: 'success',
                title: 'ส่งใบแจ้งซ่อมสำเร็จ!',
                text: `เลขที่ใบแจ้งซ่อมของคุณคือ ${data.data.ticket_no}`,
                confirmButtonColor: '#16a34a'
            });
        } else {
            Swal.fire('เกิดข้อผิดพลาด', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('เกิดข้อผิดพลาด', 'ไม่สามารถส่งข้อมูลได้', 'error');
    } finally {
        btn.disabled = false;
        btn.innerHTML = `<i class="fa-solid fa-paper-plane mr-1"></i> ยืนยันส่งใบแจ้งซ่อม`;
    }
}

// Edit Ticket Modal (Pending only)
function openEditModal(id, title, category, urgency, description) {
    document.getElementById('editTicketID').value = id;
    document.getElementById('etTitle').value = title;
    document.getElementById('etCategory').value = category;
    document.getElementById('etUrgency').value = urgency;
    document.getElementById('etDescription').value = description;
    openModal('editTicketModal');
}

async function handleUpdateTicket(e) {
    e.preventDefault();
    const id = document.getElementById('editTicketID').value;
    const req = {
        title: document.getElementById('etTitle').value,
        category: document.getElementById('etCategory').value,
        urgency: document.getElementById('etUrgency').value,
        description: document.getElementById('etDescription').value
    };

    try {
        const res = await fetch(`/api/v1/tickets/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(req)
        });
        const data = await res.json();
        if (data.success) {
            closeModal('editTicketModal');
            loadResidentTickets();
            Swal.fire({ icon: 'success', title: 'แก้ไขเรียบร้อยแล้ว', timer: 1500, showConfirmButton: false });
        } else {
            Swal.fire('ไม่สามารถแก้ไขได้', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    }
}

// Cancel Ticket (Soft Delete, Pending only)
async function confirmCancelTicket(id) {
    const result = await Swal.fire({
        title: 'ยืนยันการยกเลิกใบแจ้งซ่อม?',
        text: 'เมื่อยกเลิกแล้ว รายการนี้จะไม่ได้รับการดำเนินงานต่อ',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#e11d48',
        cancelButtonColor: '#64748b',
        confirmButtonText: 'ใช่, ยกเลิกใบแจ้งซ่อม',
        cancelButtonText: 'ย้อนกลับ'
    });

    if (result.isConfirmed) {
        try {
            const res = await fetch(`/api/v1/tickets/${id}`, { method: 'DELETE' });
            const data = await res.json();
            if (data.success) {
                loadResidentTickets();
                Swal.fire({ icon: 'success', title: 'ยกเลิกเรียบร้อย', timer: 1500, showConfirmButton: false });
            } else {
                Swal.fire('ไม่สามารถยกเลิกได้', data.message, 'error');
            }
        } catch (err) {
            Swal.fire('ข้อผิดพลาด', 'เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
        }
    }
}

// FORM 2: Feedback Rating Logic
function openFeedbackModal(ticketID) {
    document.getElementById('feedbackTicketID').value = ticketID;
    setRating(5);
    document.getElementById('fbComment').value = '';
    document.getElementById('fbTechEval').value = '';
    openModal('feedbackModal');
}

function setRating(val) {
    document.getElementById('fbRatingVal').value = val;
    const stars = document.querySelectorAll('#starRatingContainer .star-btn');
    stars.forEach((s, idx) => {
        if (idx < val) {
            s.classList.add('text-amber-400');
            s.classList.remove('text-slate-300');
        } else {
            s.classList.remove('text-amber-400');
            s.classList.add('text-slate-300');
        }
    });

    const descs = ['ต้องปรับปรุงด่วน (1 ดาว)', 'พอใช้ (2 ดาว)', 'ปานกลาง (3 ดาว)', 'ดีมาก (4 ดาว)', 'ยอดเยี่ยมมาก (5 ดาว)'];
    document.getElementById('ratingTextDesc').innerText = descs[val - 1];
}

async function handleSubmitFeedback(e) {
    e.preventDefault();
    const ticketID = document.getElementById('feedbackTicketID').value;
    const req = {
        rating: parseInt(document.getElementById('fbRatingVal').value, 10),
        comment: document.getElementById('fbComment').value,
        tech_evaluation: document.getElementById('fbTechEval').value
    };

    try {
        const res = await fetch(`/api/v1/tickets/${ticketID}/feedback`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(req)
        });
        const data = await res.json();
        if (data.success) {
            closeModal('feedbackModal');
            loadResidentTickets();
            Swal.fire({
                icon: 'success',
                title: 'บันทึกการประเมินสำเร็จ!',
                text: 'ขอบคุณสำหรับข้อเสนอแนะในการปรับปรุงการให้บริการ',
                confirmButtonColor: '#0d9488'
            });
        } else {
            Swal.fire('เกิดข้อผิดพลาด', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'ไม่สามารถส่งแบบประเมินได้', 'error');
    }
}

// ================= ADMIN & TECHNICIAN DASHBOARD =================
async function loadAdminDashboard() {
    try {
        // 1. Load Stats
        const resStats = await fetch('/api/v1/admin/dashboard');
        const dataStats = await resStats.json();
        if (dataStats.success) {
            const stats = dataStats.data;
            document.getElementById('statTotal').innerText = stats.total_tickets || 0;
            document.getElementById('statPending').innerText = stats.pending_count || 0;
            document.getElementById('statInProgress').innerText = stats.in_progress_count || 0;
            document.getElementById('statCompleted').innerText = stats.completed_count || 0;

            renderCategoryChart(stats.category_breakdown || {});
        }

        // 2. Load All Tickets
        const resTickets = await fetch('/api/v1/admin/tickets');
        const dataTickets = await resTickets.json();
        if (dataTickets.success) {
            allAdminTickets = dataTickets.data || [];
            filterAdminTickets();
        }
    } catch (err) {
        console.error('Error loading dashboard:', err);
    }
}

function renderCategoryChart(breakdown) {
    const ctx = document.getElementById('categoryChart').getContext('2d');
    const labels = ['ไฟฟ้า', 'ประปา', 'เครื่องใช้ไฟฟ้า', 'โครงสร้าง', 'อื่นๆ'];
    const keys = ['Electrical', 'Plumbing', 'Appliance', 'Structural', 'Other'];
    const dataValues = keys.map(k => breakdown[k] || 0);

    if (categoryChartInstance) {
        categoryChartInstance.destroy();
    }

    categoryChartInstance = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: [{
                data: dataValues,
                backgroundColor: ['#eab308', '#06b6d4', '#3b82f6', '#f97316', '#64748b'],
                borderWidth: 2,
                borderColor: '#ffffff'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { position: 'bottom', labels: { boxWidth: 12, font: { family: 'Prompt', size: 11 } } }
            }
        }
    });
}

function filterAdminTickets() {
    const search = document.getElementById('adminSearchInput').value.toLowerCase();
    const status = document.getElementById('adminStatusFilter').value;
    const category = document.getElementById('adminCategoryFilter').value;

    const filtered = allAdminTickets.filter(t => {
        const matchSearch = !search || 
            t.ticket_no.toLowerCase().includes(search) || 
            t.room_no.toLowerCase().includes(search) || 
            t.title.toLowerCase().includes(search) ||
            (t.user && t.user.full_name.toLowerCase().includes(search));

        const matchStatus = status === 'All' || t.status === status;
        const matchCategory = category === 'All' || t.category === category;

        return matchSearch && matchStatus && matchCategory;
    });

    renderAdminTable(filtered);
}

function renderAdminTable(tickets) {
    const tbody = document.getElementById('adminTicketTableBody');
    if (!tickets || tickets.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="text-center py-8 text-slate-400">ไม่พบข้อมูลที่ตรงกับเงื่อนไขค้นหา</td></tr>`;
        return;
    }

    let html = '';
    tickets.forEach(t => {
        const dateStr = new Date(t.created_at).toLocaleDateString('th-TH', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
        const techName = t.assigned_technician ? t.assigned_technician.full_name : '<span class="text-slate-400 italic">ยังไม่มอบหมาย</span>';
        const userName = t.user ? t.user.full_name : 'ลูกบ้าน';

        html += `
            <tr class="hover:bg-slate-50/80 transition">
                <td class="py-3 px-3">
                    <div class="font-mono font-bold text-slate-900">${t.ticket_no}</div>
                    <div class="text-[10px] text-slate-400">${dateStr}</div>
                </td>
                <td class="py-3 px-3">
                    <div class="font-bold text-slate-800">ห้อง ${t.room_no}</div>
                    <div class="text-[11px] text-slate-500">${userName}</div>
                </td>
                <td class="py-3 px-3">
                    <div class="font-semibold text-slate-800 line-clamp-1">${escapeHtml(t.title)}</div>
                    <div class="text-[11px] text-slate-500">${getCategoryName(t.category)} • ${getUrgencyBadge(t.urgency)}</div>
                </td>
                <td class="py-3 px-3">${getStatusBadge(t.status)}</td>
                <td class="py-3 px-3 text-[11px] font-medium text-slate-700">${techName}</td>
                <td class="py-3 px-3 text-right">
                    <button onclick="openAdminManageModal('${t.id}')" class="px-3 py-1.5 bg-slate-800 hover:bg-black text-white text-xs font-semibold rounded-lg shadow-sm transition">
                        จัดการงาน
                    </button>
                </td>
            </tr>`;
    });

    tbody.innerHTML = html;
}

async function loadTechniciansList() {
    try {
        const res = await fetch('/api/v1/admin/technicians');
        const data = await res.json();
        if (data.success) {
            const select = document.getElementById('admTechSelect');
            select.innerHTML = '<option value="">-- ยังไม่มอบหมายช่าง --</option>';
            data.data.forEach(tech => {
                select.innerHTML += `<option value="${tech.id}">${tech.full_name} (${tech.role})</option>`;
            });
        }
    } catch (err) {
        console.error('Error loading technicians:', err);
    }
}

function openAdminManageModal(ticketID) {
    const ticket = allAdminTickets.find(t => t.id === ticketID);
    if (!ticket) return;

    document.getElementById('adminTicketID').value = ticket.id;
    document.getElementById('admTicketNo').innerText = ticket.ticket_no;
    document.getElementById('admTicketRoomUser').innerText = `ห้อง ${ticket.room_no} (${ticket.user ? ticket.user.full_name : '-'})`;
    document.getElementById('admTicketTitle').innerText = ticket.title;
    document.getElementById('admStatusSelect').value = ticket.status;
    document.getElementById('admTechSelect').value = ticket.assigned_technician_id || '';
    document.getElementById('admInternalNote').value = ticket.internal_note || '';

    // Render uploaded photos
    const photosContainer = document.getElementById('admPhotosContainer');
    const photosGallery = document.getElementById('admPhotosGallery');
    photosGallery.innerHTML = '';

    if (ticket.images && ticket.images.length > 0) {
        photosContainer.classList.remove('hidden');
        ticket.images.forEach(img => {
            photosGallery.innerHTML += `
                <a href="${img.file_url}" target="_blank" class="w-16 h-16 rounded-lg overflow-hidden border border-slate-300 shadow-sm hover:opacity-80 transition block">
                    <img src="${img.file_url}" class="w-full h-full object-cover">
                </a>`;
        });
    } else {
        photosContainer.classList.add('hidden');
    }

    openModal('adminManageModal');
}

async function handleAdminUpdateStatus(e) {
    e.preventDefault();
    const id = document.getElementById('adminTicketID').value;
    const req = {
        status: document.getElementById('admStatusSelect').value,
        technician_id: document.getElementById('admTechSelect').value || null,
        internal_note: document.getElementById('admInternalNote').value
    };

    try {
        const res = await fetch(`/api/v1/admin/tickets/${id}/status`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(req)
        });
        const data = await res.json();
        if (data.success) {
            closeModal('adminManageModal');
            loadAdminDashboard();
            Swal.fire({ icon: 'success', title: 'อัปเดตงานซ่อมเรียบร้อย', timer: 1500, showConfirmButton: false });
        } else {
            Swal.fire('ไม่สามารถอัปเดตได้', data.message, 'error');
        }
    } catch (err) {
        Swal.fire('ข้อผิดพลาด', 'เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    }
}

// Helpers & Formatters
function getStatusBadge(status) {
    const badges = {
        'Pending': '<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-100 text-amber-800 border border-amber-200"><i class="fa-solid fa-clock text-[10px]"></i> รอดำเนินการ</span>',
        'In_Progress': '<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-800 border border-blue-200"><i class="fa-solid fa-wrench text-[10px]"></i> กำลังซ่อม</span>',
        'Completed': '<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 border border-emerald-200"><i class="fa-solid fa-check text-[10px]"></i> ซ่อมเสร็จแล้ว</span>',
        'Closed': '<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200"><i class="fa-solid fa-lock text-[10px]"></i> ปิดงานแล้ว</span>',
        'Cancelled': '<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-100 text-rose-800 border border-rose-200"><i class="fa-solid fa-xmark text-[10px]"></i> ยกเลิกแล้ว</span>'
    };
    return badges[status] || status;
}

function getUrgencyBadge(urgency) {
    const badges = {
        'Normal': '<span class="text-[11px] px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-medium">ความเร่งด่วน: ปกติ</span>',
        'Urgent': '<span class="text-[11px] px-2 py-0.5 rounded bg-amber-100 text-amber-800 font-bold">ความเร่งด่วน: ด่วน</span>',
        'Emergency': '<span class="text-[11px] px-2 py-0.5 rounded bg-rose-100 text-rose-800 font-extrabold animate-bounce">ความเร่งด่วน: ด่วนมาก</span>'
    };
    return badges[urgency] || urgency;
}

function getCategoryName(category) {
    const names = {
        'Plumbing': 'ประปา',
        'Electrical': 'ไฟฟ้า',
        'Appliance': 'เครื่องใช้ไฟฟ้า',
        'Structural': 'โครงสร้าง/อาคาร',
        'Other': 'อื่นๆ'
    };
    return names[category] || category;
}

// Modal Toggle Helpers
function openModal(id) {
    document.getElementById(id).classList.remove('hidden');
}
function closeModal(id) {
    document.getElementById(id).classList.add('hidden');
}
function switchAuthTab(tab) {
    if (tab === 'login') {
        document.getElementById('authTabLogin').className = 'tab-active pb-2';
        document.getElementById('authTabRegister').className = 'text-slate-400 pb-2';
        document.getElementById('loginForm').classList.remove('hidden');
        document.getElementById('registerForm').classList.add('hidden');
    } else {
        document.getElementById('authTabRegister').className = 'tab-active pb-2';
        document.getElementById('authTabLogin').className = 'text-slate-400 pb-2';
        document.getElementById('loginForm').classList.add('hidden');
        document.getElementById('registerForm').classList.remove('hidden');
    }
}

function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}
function escapeAttr(str) {
    if (!str) return '';
    return str.replace(/'/g, "\\'").replace(/"/g, "&quot;");
}
