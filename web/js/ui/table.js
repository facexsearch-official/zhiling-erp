// Data table component
const Table = {
  render(columns, rows, options = {}) {
    const { selectable, actions, emptyText = '暂无数据', onRowClick } = options;
    let html = '<div class="table-wrap"><table><thead><tr>';
    if (selectable) html += '<th style="width:40px"><div class="table-checkbox" onclick="Table.toggleAll(this)"></div></th>';
    columns.forEach(c => {
      const cls = c.align === 'right' ? ' class="num"' : '';
      const w = c.width ? ` style="width:${c.width}"` : '';
      html += `<th${cls}${w}>${c.label}</th>`;
    });
    if (actions) html += '<th style="width:120px">操作</th>';
    html += '</tr></thead><tbody>';

    if (!rows || rows.length === 0) {
      const colspan = columns.length + (selectable ? 1 : 0) + (actions ? 1 : 0);
      html += `<tr><td colspan="${colspan}" class="table-empty">${emptyText}</td></tr>`;
    } else {
      rows.forEach((row, idx) => {
        html += `<tr data-idx="${idx}"${onRowClick ? ` onclick="${onRowClick}(${idx})" style="cursor:pointer"` : ''}>`;
        if (selectable) html += `<td><div class="table-checkbox" onclick="event.stopPropagation();Table.toggleRow(this)"></div></td>`;
        columns.forEach(c => {
          const cls = c.align === 'right' ? ' class="num"' : '';
          let val = c.render ? c.render(row[c.key], row) : (row[c.key] ?? '-');
          html += `<td${cls}>${val}</td>`;
        });
        if (actions) html += `<td class="table-actions">${actions(row, idx)}</td>`;
        html += '</tr>';
      });
    }
    html += '</tbody></table></div>';
    return html;
  },

  pagination(total, page, pageSize, onPageChange) {
    const totalPages = Math.ceil(total / pageSize) || 1;
    let html = '<div class="table-footer"><span>共 ' + total + ' 条</span><div class="pagination">';
    html += `<button ${page <= 1 ? 'disabled' : ''} onclick="${onPageChange}(${page - 1})">&lt;</button>`;
    for (let i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= page - 2 && i <= page + 2)) {
        html += `<button class="${i === page ? 'active' : ''}" onclick="${onPageChange}(${i})">${i}</button>`;
      } else if (i === page - 3 || i === page + 3) {
        html += '<button disabled>...</button>';
      }
    }
    html += `<button ${page >= totalPages ? 'disabled' : ''} onclick="${onPageChange}(${page + 1})">&gt;</button>`;
    html += '</div></div>';
    return html;
  },

  toggleAll(el) {
    const checked = !el.classList.contains('checked');
    el.classList.toggle('checked', checked);
    el.closest('table').querySelectorAll('.table-checkbox').forEach(cb => cb.classList.toggle('checked', checked));
  },

  toggleRow(el) {
    el.classList.toggle('checked');
  },

  statusBadge(status) {
    const map = { 1: ['草稿', 'default'], 2: ['待审核', 'warning'], 3: ['已生效', 'success'], 4: ['已作废', 'danger'], 5: ['已关闭', 'default'] };
    const [text, type] = map[status] || ['未知', 'default'];
    return `<span class="badge badge-${type}">${text}</span>`;
  },
};
