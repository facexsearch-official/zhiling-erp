// Form helper
const Form = {
  getValues(container) {
    const data = {};
    container.querySelectorAll('[data-field]').forEach(el => {
      const field = el.getAttribute('data-field');
      const type = el.getAttribute('data-type') || 'string';
      let val = el.value;
      if (type === 'number') val = parseFloat(val) || 0;
      else if (type === 'int') val = parseInt(val) || 0;
      data[field] = val;
    });
    return data;
  },

  setValues(container, data) {
    container.querySelectorAll('[data-field]').forEach(el => {
      const field = el.getAttribute('data-field');
      if (data[field] !== undefined) el.value = data[field];
    });
  },

  validate(container) {
    let valid = true;
    container.querySelectorAll('[data-field][data-required]').forEach(el => {
      if (!el.value.trim()) {
        el.classList.add('error');
        valid = false;
      } else {
        el.classList.remove('error');
      }
    });
    return valid;
  },

  renderInput(label, field, opts = {}) {
    const { type = 'text', value = '', required, placeholder = '', disabled } = opts;
    const reqMark = required ? '<span class="required">*</span>' : '';
    const reqAttr = required ? ' data-required' : '';
    const disAttr = disabled ? ' disabled' : '';
    return `<div class="form-group"><label class="form-label">${label}${reqMark}</label><input class="form-input" type="${type}" data-field="${field}" value="${value}" placeholder="${placeholder}"${reqAttr}${disAttr}></div>`;
  },

  renderSelect(label, field, options, opts = {}) {
    const { value = '' } = opts;
    const reqMark = opts.required ? '<span class="required">*</span>' : '';
    let html = `<div class="form-group"><label class="form-label">${label}${reqMark}</label><select class="form-input form-select" data-field="${field}">`;
    options.forEach(o => {
      const selected = o.value == value ? ' selected' : '';
      html += `<option value="${o.value}"${selected}>${o.label}</option>`;
    });
    html += '</select></div>';
    return html;
  },

  renderTextarea(label, field, opts = {}) {
    const { value = '', placeholder = '', rows = 3 } = opts;
    return `<div class="form-group"><label class="form-label">${label}</label><textarea class="form-input form-textarea" data-field="${field}" placeholder="${placeholder}" rows="${rows}">${value}</textarea></div>`;
  },
};
