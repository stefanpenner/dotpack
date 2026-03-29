-- Leader key (must be set before plugins load)
vim.g.mapleader = " "
vim.g.maplocalleader = "\\"

-- Options
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.signcolumn = "yes"
vim.opt.cursorline = true
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2
vim.opt.smartindent = true
vim.opt.wrap = false
vim.opt.scrolloff = 8
vim.opt.sidescrolloff = 8
vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.termguicolors = true
vim.opt.splitbelow = true
vim.opt.splitright = true
vim.opt.clipboard = "unnamedplus"
vim.opt.undofile = true
vim.opt.updatetime = 200
vim.opt.timeoutlen = 300
vim.opt.completeopt = "menu,menuone,noselect"
vim.opt.pumheight = 10
vim.opt.showmode = false -- lualine shows the mode
vim.opt.fillchars = { diff = "╱", eob = " " }
vim.opt.smoothscroll = true
vim.opt.foldlevel = 99
vim.opt.grepformat = "%f:%l:%c:%m"
vim.opt.grepprg = "rg --vimgrep"

-- Disable some built-in plugins
vim.g.loaded_netrw = 1
vim.g.loaded_netrwPlugin = 1

-- Shim lazy.stats for plugins that assume lazy.nvim (e.g. snacks dashboard)
package.preload["lazy.stats"] = function()
  return {
    stats = function()
      return { startuptime = 0, loaded = 0, count = 0 }
    end,
  }
end

-- Install and load all plugins
require("plugins")

-- Configure everything
require("ui")
require("lsp")
require("treesitter")
require("editor")
require("keymaps")
