// Modal component
const Modal = {
  show(title, bodyHTML, onConfirm) {
    document.getElementById('modalTitle').textContent = title;
    document.getElementById('modalBody').innerHTML = bodyHTML;
    document.getElementById('modal').classList.add('show');
    const btn = document.getElementById('modalConfirm');
    btn.onclick = () => { if (onConfirm) onConfirm(); };
  },
  hide() {
    document.getElementById('modal').classList.remove('show');
  },
};

function showModal(title) { Modal.show(title, ''); }
function hideModal() { Modal.hide(); }
