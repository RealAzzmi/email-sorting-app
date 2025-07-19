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
  gmail_message_id: string;
  sender: string;
  subject: string;
  body: string;
  received_at: string;
}

interface PaginatedEmails {
  emails: Email[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export default function DashboardPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [emails, setEmails] = useState<Email[]>([]);
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
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

  const handleAccountSelect = (account: Account) => {
    setSelectedAccount(account);
    setSelectedEmail(null);
    setCurrentPage(1);
    fetchAccountEmails(account.id, 1);
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
        // After refresh, fetch the first page
        setCurrentPage(1);
        await fetchAccountEmails(selectedAccount.id, 1);
      } else {
        console.error('Failed to refresh emails');
      }
    } catch (error) {
      console.error('Failed to refresh emails:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const handlePageChange = (page: number) => {
    if (selectedAccount && page >= 1 && page <= totalPages) {
      setCurrentPage(page);
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

            {/* Categories Section (Placeholder) */}
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h2 className="text-lg font-medium text-black mb-4">Email Categories</h2>
              <div className="text-center py-8">
                <p className="text-gray-500 mb-4">Categories coming soon</p>
                <button
                  disabled
                  className="px-4 py-2 bg-gray-200 text-gray-400 rounded-md cursor-not-allowed"
                >
                  Add Category
                </button>
              </div>
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
                        Emails for {selectedAccount.email}
                      </h2>
                      {totalCount > 0 && (
                        <p className="text-sm text-gray-600">
                          {totalCount} emails total
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
                        <h3 className="text-xl font-medium text-black">{selectedEmail.subject}</h3>
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
                            <p className="font-medium text-black truncate">
                              {email.subject || '(No Subject)'}
                            </p>
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