'use client';

import { useState, useEffect } from 'react';

interface Account {
  id: number;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface Email {
  id: number;
  account_id: number;
  category_id?: number;
  gmail_message_id: string;
  sender: string;
  subject: string;
  body: string;
  received_at: string;
}

interface Category {
  id: number;
  account_id: number;
  name: string;
  description?: string;
}

interface PaginatedEmails {
  emails: Email[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface CreateCategoryFormProps {
  onSubmit: (name: string, description: string) => void;
  onCancel: () => void;
}

function CreateCategoryForm({ onSubmit, onCancel }: CreateCategoryFormProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim()) {
      onSubmit(name.trim(), description.trim());
      setName('');
      setDescription('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div>
        <label htmlFor="categoryName" className="block text-sm font-medium text-gray-700 mb-1">
          Category Name
        </label>
        <input
          type="text"
          id="categoryName"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-black focus:border-black"
          placeholder="Enter category name"
          required
        />
      </div>
      <div>
        <label htmlFor="categoryDescription" className="block text-sm font-medium text-gray-700 mb-1">
          Description (optional)
        </label>
        <textarea
          id="categoryDescription"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-black focus:border-black"
          placeholder="Enter description"
          rows={2}
        />
      </div>
      <div className="flex justify-end space-x-2">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-2 text-sm border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-3 py-2 text-sm bg-black text-white rounded-md hover:bg-gray-800 transition-colors"
        >
          Create
        </button>
      </div>
    </form>
  );
}

export default function DashboardPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [emails, setEmails] = useState<Email[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [showCreateCategoryForm, setShowCreateCategoryForm] = useState(false);
  const [showCustomDropdown, setShowCustomDropdown] = useState(false);
  const [loading, setLoading] = useState(true);
  const [emailsLoading, setEmailsLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const pageSize = 20;

  useEffect(() => {
    fetchAccounts();
  }, []);

  const fetchAccounts = async () => {
    try {
      const response = await fetch('http://localhost:8080/accounts', {
        credentials: 'include',
      });
      const data = await response.json();
      setAccounts(data.accounts || []);
    } catch (error) {
      console.error('Failed to fetch accounts:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchAccountEmails = async (accountId: number, page: number = 1) => {
    setEmailsLoading(true);
    try {
      const response = await fetch(`http://localhost:8080/accounts/${accountId}/emails?page=${page}&page_size=${pageSize}`, {
        credentials: 'include',
      });
      const data: PaginatedEmails = await response.json();
      setEmails(data.emails || []);
      setCurrentPage(data.page);
      setTotalPages(data.total_pages);
      setTotalCount(data.total_count);
    } catch (error) {
      console.error('Failed to fetch emails:', error);
      setEmails([]);
      setCurrentPage(1);
      setTotalPages(1);
      setTotalCount(0);
    } finally {
      setEmailsLoading(false);
    }
  };

  const fetchAccountCategories = async (accountId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/accounts/${accountId}/categories`, {
        credentials: 'include',
      });
      const data = await response.json();
      setCategories(data.categories || []);
    } catch (error) {
      console.error('Failed to fetch categories:', error);
      setCategories([]);
    }
  };

  const handleAccountSelect = (account: Account) => {
    setSelectedAccount(account);
    setSelectedEmail(null);
    setSelectedCategory(null);
    setCurrentPage(1);
    fetchAccountEmails(account.id, 1);
    fetchAccountCategories(account.id);
  };

  const getCategoryName = (categoryId?: number): string => {
    if (!categoryId) return '';
    const category = categories.find(c => c.id === categoryId);
    return category ? category.name : '';
  };

  const isSystemLabel = (categoryName: string): boolean => {
    const systemLabels = [
      'Inbox', 'Sent', 'Drafts', 'Spam', 'Trash', 'Important', 'Starred', 'All Mail', 'Chats',
      'INBOX', 'SENT', 'DRAFT', 'SPAM', 'TRASH', 'IMPORTANT', 'STARRED', 'UNREAD', 'CHAT'
    ];
    return systemLabels.includes(categoryName);
  };

  const getSystemCategories = (): Category[] => {
    return categories.filter(cat => isSystemLabel(cat.name));
  };

  const getCustomCategories = (): Category[] => {
    return categories.filter(cat => !isSystemLabel(cat.name));
  };

  const handleCategorySelect = async (category: Category) => {
    if (!selectedAccount) return;
    
    setSelectedCategory(category);
    setSelectedEmail(null);
    setCurrentPage(1);
    
    // Fetch emails for this category
    setEmailsLoading(true);
    try {
      const response = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/categories/${category.id}/emails?page=1&page_size=${pageSize}`, {
        credentials: 'include',
      });
      const data: PaginatedEmails = await response.json();
      setEmails(data.emails || []);
      setCurrentPage(data.page);
      setTotalPages(data.total_pages);
      setTotalCount(data.total_count);
    } catch (error) {
      console.error('Failed to fetch emails by category:', error);
      setEmails([]);
      setCurrentPage(1);
      setTotalPages(1);
      setTotalCount(0);
    } finally {
      setEmailsLoading(false);
    }
  };

  const handleCreateCategory = async (name: string, description: string) => {
    if (!selectedAccount) return;
    
    try {
      const response = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/categories`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ name, description }),
      });
      
      if (response.ok) {
        // Refresh categories list
        await fetchAccountCategories(selectedAccount.id);
        setShowCreateCategoryForm(false);
      } else {
        console.error('Failed to create category');
      }
    } catch (error) {
      console.error('Failed to create category:', error);
    }
  };

  const handleDeleteCategory = async (categoryId: number, event: React.MouseEvent) => {
    event.stopPropagation();
    if (!selectedAccount) return;
    
    if (confirm('Are you sure you want to delete this category?')) {
      try {
        const response = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/categories/${categoryId}`, {
          method: 'DELETE',
          credentials: 'include',
        });
        
        if (response.ok) {
          // Refresh categories list
          await fetchAccountCategories(selectedAccount.id);
          // If this was the selected category, clear selection
          if (selectedCategory?.id === categoryId) {
            setSelectedCategory(null);
            fetchAccountEmails(selectedAccount.id, 1);
          }
        } else {
          console.error('Failed to delete category');
        }
      } catch (error) {
        console.error('Failed to delete category:', error);
      }
    }
  };

  const handleBackToAllEmails = () => {
    if (!selectedAccount) return;
    
    setSelectedCategory(null);
    setCurrentPage(1);
    fetchAccountEmails(selectedAccount.id, 1);
  };

  const handleRefreshEmails = async () => {
    if (!selectedAccount) return;
    
    setRefreshing(true);
    try {
      const response = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/emails/refresh`, {
        method: 'POST',
        credentials: 'include',
      });
      
      if (response.ok) {
        // After refresh, fetch the first page and categories
        setCurrentPage(1);
        await Promise.all([
          fetchAccountEmails(selectedAccount.id, 1),
          fetchAccountCategories(selectedAccount.id)
        ]);
      } else {
        console.error('Failed to refresh emails');
      }
    } catch (error) {
      console.error('Failed to refresh emails:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const handlePageChange = async (page: number) => {
    if (!selectedAccount || page < 1 || page > totalPages) return;
    
    setCurrentPage(page);
    
    if (selectedCategory) {
      // Fetch emails for the selected category
      setEmailsLoading(true);
      try {
        const response = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/categories/${selectedCategory.id}/emails?page=${page}&page_size=${pageSize}`, {
          credentials: 'include',
        });
        const data: PaginatedEmails = await response.json();
        setEmails(data.emails || []);
        setCurrentPage(data.page);
        setTotalPages(data.total_pages);
        setTotalCount(data.total_count);
      } catch (error) {
        console.error('Failed to fetch emails by category:', error);
      } finally {
        setEmailsLoading(false);
      }
    } else {
      // Fetch all emails
      fetchAccountEmails(selectedAccount.id, page);
    }
  };

  const handleConnectNewAccount = () => {
    window.location.href = '/login';
  };

  const handleLogout = async () => {
    try {
      await fetch('http://localhost:8080/auth/logout', {
        method: 'POST',
        credentials: 'include',
      });
      window.location.href = '/login';
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleDeleteAccount = async (accountId: number, event: React.MouseEvent) => {
    event.stopPropagation();
    if (confirm('Are you sure you want to remove this account?')) {
      try {
        await fetch(`http://localhost:8080/accounts/${accountId}`, {
          method: 'DELETE',
          credentials: 'include',
        });
        setAccounts(accounts.filter(acc => acc.id !== accountId));
        if (selectedAccount?.id === accountId) {
          setSelectedAccount(null);
          setEmails([]);
        }
      } catch (error) {
        console.error('Failed to delete account:', error);
      }
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-gray-300 border-t-black rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold text-black">Email Sorting App</h1>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm border border-gray-300 rounded-md text-black hover:bg-gray-50 transition-colors"
            >
              Sign Out
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Left Column: Connected Accounts */}
          <div className="space-y-6">
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h2 className="text-lg font-medium text-black mb-4">Connected Gmail Accounts</h2>
              
              {accounts.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-gray-500 mb-4">No accounts connected yet</p>
                  <button
                    onClick={handleConnectNewAccount}
                    className="px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800 transition-colors"
                  >
                    Connect Gmail Account
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  {accounts.map((account) => (
                    <div
                      key={account.id}
                      onClick={() => handleAccountSelect(account)}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                        selectedAccount?.id === account.id
                          ? 'border-black bg-gray-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <p className="font-medium text-black">{account.name}</p>
                          <p className="text-sm text-gray-600">{account.email}</p>
                        </div>
                        <button
                          onClick={(e) => handleDeleteAccount(account.id, e)}
                          className="text-gray-400 hover:text-red-500 transition-colors"
                        >
                          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"/>
                          </svg>
                        </button>
                      </div>
                    </div>
                  ))}
                  
                  <button
                    onClick={handleConnectNewAccount}
                    className="w-full p-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-gray-400 hover:text-gray-700 transition-colors"
                  >
                    + Connect Another Account
                  </button>
                </div>
              )}
            </div>

            {/* Categories Section */}
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-medium text-black">Email Categories</h2>
                <button
                  onClick={() => setShowCreateCategoryForm(true)}
                  className="px-3 py-1 text-sm bg-black text-white rounded-md hover:bg-gray-800 transition-colors"
                >
                  + Add
                </button>
              </div>
              
              {showCreateCategoryForm && (
                <div className="mb-4 p-4 border border-gray-200 rounded-lg bg-gray-50">
                  <CreateCategoryForm 
                    onSubmit={handleCreateCategory}
                    onCancel={() => setShowCreateCategoryForm(false)}
                  />
                </div>
              )}
              
              {categories.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-gray-500">No categories yet</p>
                </div>
              ) : (
                <div className="space-y-2">
                  <button
                    onClick={handleBackToAllEmails}
                    className={`w-full text-left p-3 rounded-lg transition-colors ${
                      !selectedCategory 
                        ? 'bg-black text-white' 
                        : 'hover:bg-gray-50 text-gray-700'
                    }`}
                  >
                    All Emails ({totalCount})
                  </button>
                  
                  {/* System Categories */}
                  {getSystemCategories().map((category) => (
                    <div
                      key={category.id}
                      onClick={() => handleCategorySelect(category)}
                      className={`flex justify-between items-center p-3 rounded-lg cursor-pointer transition-colors ${
                        selectedCategory?.id === category.id
                          ? 'bg-black text-white'
                          : 'hover:bg-gray-50 text-gray-700'
                      }`}
                    >
                      <div className="flex-1">
                        <p className="font-medium">{category.name}</p>
                      </div>
                    </div>
                  ))}
                  
                  {/* Custom Categories Dropdown */}
                  {getCustomCategories().length > 0 && (
                    <div className="relative">
                      <button
                        onClick={() => setShowCustomDropdown(!showCustomDropdown)}
                        className="w-full text-left p-3 rounded-lg hover:bg-gray-50 text-gray-700 transition-colors flex justify-between items-center"
                      >
                        <span className="font-medium">Custom Labels ({getCustomCategories().length})</span>
                        <svg 
                          className={`w-4 h-4 transition-transform ${showCustomDropdown ? 'rotate-180' : ''}`} 
                          fill="currentColor" 
                          viewBox="0 0 20 20"
                        >
                          <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd"/>
                        </svg>
                      </button>
                      
                      {showCustomDropdown && (
                        <div className="mt-2 ml-4 space-y-1 border-l-2 border-gray-200 pl-3">
                          {getCustomCategories().map((category) => (
                            <div
                              key={category.id}
                              onClick={() => {
                                handleCategorySelect(category);
                                setShowCustomDropdown(false);
                              }}
                              className={`flex justify-between items-center p-2 rounded-lg cursor-pointer transition-colors ${
                                selectedCategory?.id === category.id
                                  ? 'bg-black text-white'
                                  : 'hover:bg-gray-50 text-gray-700'
                              }`}
                            >
                              <div className="flex-1">
                                <p className="font-medium text-sm">{category.name}</p>
                                {category.description && (
                                  <p className="text-xs opacity-75">{category.description}</p>
                                )}
                              </div>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleDeleteCategory(category.id, e);
                                }}
                                className="opacity-50 hover:opacity-100 transition-opacity ml-2"
                              >
                                <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                  <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"/>
                                </svg>
                              </button>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Right Column: Email Content */}
          <div className="lg:col-span-2">
            {selectedAccount ? (
              <div className="bg-white border border-gray-200 rounded-lg">
                <div className="border-b border-gray-200 px-6 py-4">
                  <div className="flex justify-between items-center">
                    <div>
                      <h2 className="text-lg font-medium text-black">
                        {selectedCategory 
                          ? `${selectedCategory.name} - ${selectedAccount.email}`
                          : `Emails for ${selectedAccount.email}`
                        }
                      </h2>
                      {totalCount > 0 && (
                        <p className="text-sm text-gray-600">
                          {totalCount} {selectedCategory ? `emails in ${selectedCategory.name}` : 'emails total'}
                        </p>
                      )}
                    </div>
                    <button
                      onClick={handleRefreshEmails}
                      disabled={refreshing || emailsLoading}
                      className="px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                      {refreshing ? (
                        <div className="flex items-center">
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                          Refreshing...
                        </div>
                      ) : (
                        'Refresh'
                      )}
                    </button>
                  </div>
                </div>
                
                {emailsLoading ? (
                  <div className="p-6 text-center">
                    <div className="w-6 h-6 border-2 border-gray-300 border-t-black rounded-full animate-spin mx-auto mb-2"></div>
                    <p className="text-gray-600">Loading emails...</p>
                  </div>
                ) : emails.length === 0 ? (
                  <div className="p-6 text-center">
                    <p className="text-gray-500">No emails found</p>
                  </div>
                ) : selectedEmail ? (
                  <div className="p-6">
                    <button
                      onClick={() => setSelectedEmail(null)}
                      className="mb-4 text-sm text-gray-600 hover:text-black transition-colors"
                    >
                      ← Back to email list
                    </button>
                    <div className="space-y-4">
                      <div>
                        <div className="flex items-center gap-2 mb-2">
                          <h3 className="text-xl font-medium text-black">{selectedEmail.subject}</h3>
                          {selectedEmail.category_id && (
                            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                              {getCategoryName(selectedEmail.category_id)}
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600">From: {selectedEmail.sender}</p>
                        <p className="text-sm text-gray-600">
                          Received: {new Date(selectedEmail.received_at).toLocaleString()}
                        </p>
                      </div>
                      <div className="border-t border-gray-200 pt-4">
                        <pre className="whitespace-pre-wrap text-sm text-gray-700 bg-gray-50 p-4 rounded border">
                          {selectedEmail.body || 'No content available'}
                        </pre>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="divide-y divide-gray-200">
                    {emails.map((email) => (
                      <div
                        key={email.id}
                        onClick={() => setSelectedEmail(email)}
                        className="p-4 hover:bg-gray-50 cursor-pointer transition-colors"
                      >
                        <div className="flex justify-between items-start">
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <p className="font-medium text-black truncate">
                                {email.subject || '(No Subject)'}
                              </p>
                              {email.category_id && (
                                <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                  {getCategoryName(email.category_id)}
                                </span>
                              )}
                            </div>
                            <p className="text-sm text-gray-600 truncate">
                              {email.sender}
                            </p>
                            <p className="text-xs text-gray-500">
                              {new Date(email.received_at).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                
                {/* Pagination */}
                {!emailsLoading && !selectedEmail && emails.length > 0 && totalPages > 1 && (
                  <div className="border-t border-gray-200 px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="text-sm text-gray-600">
                        Page {currentPage} of {totalPages}
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handlePageChange(currentPage - 1)}
                          disabled={currentPage <= 1}
                          className="px-3 py-1 border border-gray-300 rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
                        >
                          Previous
                        </button>
                        
                        {/* Page numbers */}
                        {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                          const startPage = Math.max(1, currentPage - 2);
                          const pageNum = startPage + i;
                          if (pageNum > totalPages) return null;
                          
                          return (
                            <button
                              key={pageNum}
                              onClick={() => handlePageChange(pageNum)}
                              className={`px-3 py-1 text-sm rounded transition-colors ${
                                pageNum === currentPage
                                  ? 'bg-black text-white'
                                  : 'border border-gray-300 hover:bg-gray-50'
                              }`}
                            >
                              {pageNum}
                            </button>
                          );
                        })}
                        
                        <button
                          onClick={() => handlePageChange(currentPage + 1)}
                          disabled={currentPage >= totalPages}
                          className="px-3 py-1 border border-gray-300 rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
                        >
                          Next
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <div className="text-center py-12">
                  <p className="text-gray-500">Select a Gmail account to view emails</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}