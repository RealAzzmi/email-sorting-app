'use client';

import { useState, useEffect, useCallback } from 'react';
import DOMPurify from 'dompurify';

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
  category_ids: number[];
  categories?: Category[];
  gmail_message_id: string;
  sender: string;
  subject: string;
  body: string;
  ai_summary?: string;
  unsubscribe_link?: string;
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
          className="w-full px-3 py-2 bg-white rounded-md focus:outline-none focus:ring-2 focus:ring-black shadow-sm"
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
          className="w-full px-3 py-2 bg-white rounded-md focus:outline-none focus:ring-2 focus:ring-black shadow-sm"
          placeholder="Enter description"
          rows={2}
        />
      </div>
      <div className="flex justify-end space-x-2">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-2 text-sm bg-gray-100 rounded-md text-gray-700 hover:bg-gray-200 transition-colors"
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
  const [showAISummary, setShowAISummary] = useState(false);
  const [emailsShowingAISummary, setEmailsShowingAISummary] = useState<Set<number>>(new Set());
  const [selectedEmailIds, setSelectedEmailIds] = useState<Set<number>>(new Set());
  const [isUnsubscribing, setIsUnsubscribing] = useState(false);
  const [isSummarizing, setIsSummarizing] = useState(false);
  const [isCategorizing, setIsCategorizing] = useState(false);
  const pageSize = 20;

  const fetchAccounts = useCallback(async () => {
    try {
      const response = await fetch('http://localhost:8080/accounts', {
        credentials: 'include',
      });
      const data = await response.json();
      const accounts = data.accounts || [];
      setAccounts(accounts);
      
      // Automatically select the first account if there are any accounts
      if (accounts.length > 0 && !selectedAccount) {
        const firstAccount = accounts[0];
        setSelectedAccount(firstAccount);
        setSelectedEmail(null);
        setSelectedCategory(null);
        setCurrentPage(1);
        fetchAccountEmails(firstAccount.id, 1);
        fetchAccountCategories(firstAccount.id);
      }
    } catch (error) {
      console.error('Failed to fetch accounts:', error);
    } finally {
      setLoading(false);
    }
  }, [selectedAccount]);

  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

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
    setSelectedEmailIds(new Set());
    setEmailsShowingAISummary(new Set());
    setCurrentPage(1);
    fetchAccountEmails(account.id, 1);
    fetchAccountCategories(account.id);
  };

  const getCategoryNames = (categoryIds: number[]): string[] => {
    if (!categoryIds || categoryIds.length === 0) return [];
    return categoryIds.map(id => {
      const category = categories.find(c => c.id === id);
      return category ? category.name : '';
    }).filter(name => name !== '');
  };

  const getNonSystemCategoryNames = (categoryIds: number[]): string[] => {
    if (!categoryIds || categoryIds.length === 0) return [];
    return categoryIds.map(id => {
      const category = categories.find(c => c.id === id);
      return category ? category.name : '';
    }).filter(name => name !== '' && !isSystemLabel(name));
  };

  const isSystemLabel = (categoryName: string): boolean => {
    const systemLabels = [
      // Friendly names
      'Inbox', 'Sent', 'Drafts', 'Spam', 'Trash', 'Important', 'Starred', 'All Mail', 'Chats',
      // Gmail system label IDs
      'INBOX', 'SENT', 'DRAFT', 'SPAM', 'TRASH', 'IMPORTANT', 'STARRED', 'UNREAD', 'CHAT',
      // Star labels
      'YELLOW_STAR', 'BLUE_STAR', 'RED_STAR', 'ORANGE_STAR', 'GREEN_STAR', 'PURPLE_STAR',
      // Category labels
      'CATEGORY_PERSONAL', 'CATEGORY_SOCIAL', 'CATEGORY_PROMOTIONS', 'CATEGORY_UPDATES', 'CATEGORY_FORUMS'
    ];
    return systemLabels.includes(categoryName);
  };

  const getSystemCategories = (): Category[] => {
    return categories.filter(cat => isSystemLabel(cat.name));
  };

  const getCustomCategories = (): Category[] => {
    return categories.filter(cat => !isSystemLabel(cat.name));
  };

  const sanitizeAndRenderEmailBody = (body: string): string => {
    if (!body) return '';
    
    const isHTML = body.includes('<') && body.includes('>');
    
    if (isHTML) {
      // Sanitize HTML content with more permissive styling for emails
      return DOMPurify.sanitize(body, {
        ALLOWED_TAGS: [
          'p', 'div', 'span', 'br', 'strong', 'b', 'em', 'i', 'u', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
          'ul', 'ol', 'li', 'blockquote', 'pre', 'code', 'a', 'img', 'table', 'thead', 'tbody', 'tr', 'td', 'th',
          'center', 'font', 'small', 'big', 'sup', 'sub', 'hr', 'dl', 'dt', 'dd'
        ],
        ALLOWED_ATTR: [
          'href', 'target', 'rel', 'src', 'alt', 'width', 'height', 'style', 'class', 'id',
          'align', 'valign', 'bgcolor', 'color', 'size', 'face', 'border', 'cellpadding', 'cellspacing',
          'marginwidth', 'marginheight', 'leftmargin', 'topmargin', 'rightmargin', 'bottommargin'
        ],
        ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
        ADD_ATTR: ['target'],
        ADD_DATA_URI_TAGS: ['img'],
        FORCE_BODY: false,
        KEEP_CONTENT: true,
        ALLOW_DATA_ATTR: false
      });
    } else {
      // For plain text, escape HTML and preserve whitespace
      const escaped = body
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#x27;');
      
      return `<pre style="white-space: pre-wrap; font-family: inherit; margin: 0;">${escaped}</pre>`;
    }
  };

  const handleCategorySelect = async (category: Category) => {
    if (!selectedAccount) return;
    
    setSelectedCategory(category);
    setSelectedEmail(null);
    setSelectedEmailIds(new Set());
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
    setSelectedEmailIds(new Set());
    setEmailsShowingAISummary(new Set());
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
        
        // Maintain the current view (all emails or specific category)
        if (selectedCategory) {
          // If viewing a specific category, fetch that category's emails
          setEmailsLoading(true);
          try {
            const emailResponse = await fetch(`http://localhost:8080/accounts/${selectedAccount.id}/categories/${selectedCategory.id}/emails?page=1&page_size=${pageSize}`, {
              credentials: 'include',
            });
            const emailData = await emailResponse.json();
            setEmails(emailData.emails || []);
            setCurrentPage(emailData.page);
            setTotalPages(emailData.total_pages);
            setTotalCount(emailData.total_count);
          } catch (error) {
            console.error('Failed to fetch emails by category after refresh:', error);
          } finally {
            setEmailsLoading(false);
          }
          
          // Also refresh categories
          await fetchAccountCategories(selectedAccount.id);
        } else {
          // If viewing all emails, fetch all emails
          await Promise.all([
            fetchAccountEmails(selectedAccount.id, 1),
            fetchAccountCategories(selectedAccount.id)
          ]);
        }
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

  const handleGenerateAISummary = async (emailId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/emails/${emailId}/summary`, {
        method: 'POST',
        credentials: 'include',
      });
      
      if (response.ok) {
        // Refresh the current view to get the updated email with AI summary
        if (selectedCategory) {
          const emailResponse = await fetch(`http://localhost:8080/accounts/${selectedAccount!.id}/categories/${selectedCategory.id}/emails?page=${currentPage}&page_size=${pageSize}`, {
            credentials: 'include',
          });
          const emailData = await emailResponse.json();
          setEmails(emailData.emails || []);
          
          // Update selected email if it's the one we generated summary for
          if (selectedEmail && selectedEmail.id === emailId) {
            const updatedEmail = emailData.emails.find((e: Email) => e.id === emailId);
            if (updatedEmail) {
              setSelectedEmail(updatedEmail);
            }
          }
        } else {
          fetchAccountEmails(selectedAccount!.id, currentPage);
          
          // Update selected email if it's the one we generated summary for
          if (selectedEmail && selectedEmail.id === emailId) {
            const emailResponse = await fetch(`http://localhost:8080/accounts/${selectedAccount!.id}/emails?page=${currentPage}&page_size=${pageSize}`, {
              credentials: 'include',
            });
            const emailData = await emailResponse.json();
            const updatedEmail = emailData.emails.find((e: Email) => e.id === emailId);
            if (updatedEmail) {
              setSelectedEmail(updatedEmail);
            }
          }
        }
      } else {
        console.error('Failed to generate AI summary');
      }
    } catch (error) {
      console.error('Failed to generate AI summary:', error);
    }
  };

  const handleCategorizeWithAI = async (emailId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/emails/${emailId}/categorize`, {
        method: 'POST',
        credentials: 'include',
      });
      
      if (response.ok) {
        // Refresh the current view to get the updated email with new categories
        if (selectedCategory) {
          const emailResponse = await fetch(`http://localhost:8080/accounts/${selectedAccount!.id}/categories/${selectedCategory.id}/emails?page=${currentPage}&page_size=${pageSize}`, {
            credentials: 'include',
          });
          const emailData = await emailResponse.json();
          setEmails(emailData.emails || []);
          
          // Update selected email if it's the one we categorized
          if (selectedEmail && selectedEmail.id === emailId) {
            const updatedEmail = emailData.emails.find((e: Email) => e.id === emailId);
            if (updatedEmail) {
              setSelectedEmail(updatedEmail);
            }
          }
        } else {
          fetchAccountEmails(selectedAccount!.id, currentPage);
          
          // Update selected email if it's the one we categorized
          if (selectedEmail && selectedEmail.id === emailId) {
            const emailResponse = await fetch(`http://localhost:8080/accounts/${selectedAccount!.id}/emails?page=${currentPage}&page_size=${pageSize}`, {
              credentials: 'include',
            });
            const emailData = await emailResponse.json();
            const updatedEmail = emailData.emails.find((e: Email) => e.id === emailId);
            if (updatedEmail) {
              setSelectedEmail(updatedEmail);
            }
          }
        }
        
        // Also refresh categories in case new ones were created
        await fetchAccountCategories(selectedAccount!.id);
      } else {
        const errorData = await response.json();
        console.error('Failed to categorize email:', errorData.error);
        alert(errorData.error || 'Failed to categorize email');
      }
    } catch (error) {
      console.error('Failed to categorize email:', error);
      alert('Failed to categorize email');
    }
  };

  const handleSelectEmail = (emailId: number) => {
    const newSelected = new Set(selectedEmailIds);
    if (newSelected.has(emailId)) {
      newSelected.delete(emailId);
    } else {
      newSelected.add(emailId);
    }
    setSelectedEmailIds(newSelected);
  };

  const handleToggleEmailAISummary = (emailId: number, email: Email) => {
    const newShowingSummary = new Set(emailsShowingAISummary);
    
    if (email.ai_summary) {
      // If email has a summary, toggle its display
      if (newShowingSummary.has(emailId)) {
        newShowingSummary.delete(emailId);
      } else {
        newShowingSummary.add(emailId);
      }
      setEmailsShowingAISummary(newShowingSummary);
    } else {
      // If email doesn't have a summary, generate one and show it
      handleGenerateAISummary(emailId);
      newShowingSummary.add(emailId);
      setEmailsShowingAISummary(newShowingSummary);
    }
  };

  const handleSelectAll = () => {
    if (selectedEmailIds.size === emails.length) {
      setSelectedEmailIds(new Set());
    } else {
      setSelectedEmailIds(new Set(emails.map(email => email.id)));
    }
  };

  const handleSingleUnsubscribe = async (emailId: number) => {
    if (!confirm('Are you sure you want to unsubscribe from this email?')) {
      return;
    }

    try {
      const response = await fetch(`http://localhost:8080/emails/${emailId}/unsubscribe`, {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success) {
          alert('Successfully unsubscribed!');
        } else {
          let errorMsg = 'Unsubscribe failed';
          if (result.message) {
            errorMsg += `: ${result.message}`;
          }
          if (result.error_type) {
            errorMsg += ` (${result.error_type})`;
          }
          alert(errorMsg);
        }
        
        // Refresh emails view
        if (selectedCategory) {
          handleCategorySelect(selectedCategory);
        } else {
          fetchAccountEmails(selectedAccount!.id, currentPage);
        }
      } else {
        const errorData = await response.json();
        console.error('Failed to unsubscribe:', errorData.error);
        alert('Failed to unsubscribe: ' + (errorData.error || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to unsubscribe:', error);
      alert('Failed to unsubscribe: Network error');
    }
  };

  const handleBulkUnsubscribe = async () => {
    if (selectedEmailIds.size === 0) {
      alert('Please select emails to unsubscribe from');
      return;
    }

    if (!confirm(`Are you sure you want to unsubscribe from ${selectedEmailIds.size} emails?`)) {
      return;
    }

    setIsUnsubscribing(true);
    try {
      const response = await fetch('http://localhost:8080/emails/bulk-unsubscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          email_ids: Array.from(selectedEmailIds)
        }),
      });

      if (response.ok) {
        const data = await response.json();
        const results = data.results;
        
        // Count successful and failed unsubscribes and collect error details
        let successCount = 0;
        let failureCount = 0;
        const failureMessages: string[] = [];
        
        Object.entries(results as Record<string, {success: boolean, message?: string, error_type?: string}>).forEach(([emailId, result]) => {
          if (result.success) {
            successCount++;
          } else {
            failureCount++;
            if (result.message) {
              failureMessages.push(`Email ${emailId}: ${result.message}`);
            } else {
              failureMessages.push(`Email ${emailId}: Unknown error`);
            }
          }
        });

        // Show detailed results
        let message = `Unsubscribe completed: ${successCount} successful, ${failureCount} failed`;
        if (failureCount > 0 && failureMessages.length > 0) {
          message += '\n\nFailures:\n' + failureMessages.slice(0, 5).join('\n');
          if (failureMessages.length > 5) {
            message += `\n... and ${failureMessages.length - 5} more`;
          }
        }
        alert(message);
        
        // Clear selection
        setSelectedEmailIds(new Set());
        
        // Refresh emails view
        if (selectedCategory) {
          handleCategorySelect(selectedCategory);
        } else {
          fetchAccountEmails(selectedAccount!.id, currentPage);
        }
      } else {
        const errorData = await response.json();
        console.error('Failed to bulk unsubscribe:', errorData.error);
        alert('Failed to unsubscribe: ' + (errorData.error || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to bulk unsubscribe:', error);
      alert('Failed to unsubscribe: Network error');
    } finally {
      setIsUnsubscribing(false);
    }
  };

  const retryWithBackoff = async (
    fn: () => Promise<Response>,
    maxRetries: number = 3,
    baseDelay: number = 1000
  ): Promise<Response> => {
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const response = await fn();
        if (response.ok || response.status !== 429) {
          return response;
        }
        
        if (attempt === maxRetries) {
          return response;
        }
        
        // Exponential backoff with jitter
        const delay = baseDelay * Math.pow(2, attempt) + Math.random() * 1000;
        await new Promise(resolve => setTimeout(resolve, delay));
      } catch (error) {
        if (attempt === maxRetries) {
          throw error;
        }
        
        const delay = baseDelay * Math.pow(2, attempt) + Math.random() * 1000;
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }
    
    throw new Error('Max retries exceeded');
  };

  const handleBulkSummarize = async () => {
    if (selectedEmailIds.size === 0) {
      alert('Please select emails to summarize');
      return;
    }

    if (!confirm(`Are you sure you want to generate AI summaries for ${selectedEmailIds.size} emails?`)) {
      return;
    }

    setIsSummarizing(true);
    try {
      const emailIds = Array.from(selectedEmailIds);
      const results = await Promise.allSettled(
        emailIds.map(async (emailId) => {
          try {
            const response = await retryWithBackoff(() =>
              fetch(`http://localhost:8080/emails/${emailId}/summary`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                },
                credentials: 'include',
              })
            );

            if (response.ok) {
              return { emailId, success: true, message: 'Summary generated' };
            } else {
              const errorData = await response.json();
              return { 
                emailId, 
                success: false, 
                message: errorData.error || 'Failed to generate summary' 
              };
            }
          } catch (error) {
            return { 
              emailId, 
              success: false, 
              message: error instanceof Error ? error.message : 'Network error' 
            };
          }
        })
      );

      // Process results
      let successCount = 0;
      let failureCount = 0;
      const failureMessages: string[] = [];

      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          if (result.value.success) {
            successCount++;
          } else {
            failureCount++;
            failureMessages.push(`Email ${result.value.emailId}: ${result.value.message}`);
          }
        } else {
          failureCount++;
          failureMessages.push(`Email ${emailIds[index]}: ${result.reason}`);
        }
      });

      // Show results
      let message = `Summarization completed: ${successCount} successful, ${failureCount} failed`;
      if (failureCount > 0 && failureMessages.length > 0) {
        message += '\n\nFailures:\n' + failureMessages.slice(0, 5).join('\n');
        if (failureMessages.length > 5) {
          message += `\n... and ${failureMessages.length - 5} more`;
        }
      }
      alert(message);

      // Clear selection and refresh
      setSelectedEmailIds(new Set());
      if (selectedCategory) {
        handleCategorySelect(selectedCategory);
      } else {
        fetchAccountEmails(selectedAccount!.id, currentPage);
      }
    } catch (error) {
      console.error('Failed to bulk summarize:', error);
      alert('Failed to summarize: Network error');
    } finally {
      setIsSummarizing(false);
    }
  };

  const handleBulkCategorize = async () => {
    if (selectedEmailIds.size === 0) {
      alert('Please select emails to categorize');
      return;
    }

    if (!confirm(`Are you sure you want to AI-categorize ${selectedEmailIds.size} emails?`)) {
      return;
    }

    setIsCategorizing(true);
    try {
      const emailIds = Array.from(selectedEmailIds);
      const results = await Promise.allSettled(
        emailIds.map(async (emailId) => {
          try {
            const response = await retryWithBackoff(() =>
              fetch(`http://localhost:8080/emails/${emailId}/categorize`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                },
                credentials: 'include',
              })
            );

            if (response.ok) {
              const data = await response.json();
              return { 
                emailId, 
                success: true, 
                message: `Categorized as: ${data.categories?.join(', ') || 'Unknown'}` 
              };
            } else {
              const errorData = await response.json();
              return { 
                emailId, 
                success: false, 
                message: errorData.error || 'Failed to categorize' 
              };
            }
          } catch (error) {
            return { 
              emailId, 
              success: false, 
              message: error instanceof Error ? error.message : 'Network error' 
            };
          }
        })
      );

      // Process results
      let successCount = 0;
      let failureCount = 0;
      const failureMessages: string[] = [];

      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          if (result.value.success) {
            successCount++;
          } else {
            failureCount++;
            failureMessages.push(`Email ${result.value.emailId}: ${result.value.message}`);
          }
        } else {
          failureCount++;
          failureMessages.push(`Email ${emailIds[index]}: ${result.reason}`);
        }
      });

      // Show results
      let message = `Categorization completed: ${successCount} successful, ${failureCount} failed`;
      if (failureCount > 0 && failureMessages.length > 0) {
        message += '\n\nFailures:\n' + failureMessages.slice(0, 5).join('\n');
        if (failureMessages.length > 5) {
          message += `\n... and ${failureMessages.length - 5} more`;
        }
      }
      alert(message);

      // Clear selection and refresh
      setSelectedEmailIds(new Set());
      if (selectedCategory) {
        handleCategorySelect(selectedCategory);
      } else {
        fetchAccountEmails(selectedAccount!.id, currentPage);
      }
    } catch (error) {
      console.error('Failed to bulk categorize:', error);
      alert('Failed to categorize: Network error');
    } finally {
      setIsCategorizing(false);
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
    <>
      <style jsx global>{`
        html {
          scrollbar-gutter: stable;
        }
      `}</style>
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold text-black">Email Sorting App</h1>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm bg-gray-100 rounded-md text-black hover:bg-gray-200 transition-colors"
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
            <div className="bg-white rounded-lg p-6 shadow-sm">
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
                      className={`p-3 rounded-lg cursor-pointer transition-colors ${
                        selectedAccount?.id === account.id
                          ? 'bg-black text-white'
                          : 'bg-gray-50 hover:bg-gray-100'
                      }`}
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <p className={`font-medium ${selectedAccount?.id === account.id ? 'text-white' : 'text-black'}`}>{account.name}</p>
                          <p className={`text-sm ${selectedAccount?.id === account.id ? 'text-gray-200' : 'text-gray-600'}`}>{account.email}</p>
                        </div>
                        <button
                          onClick={(e) => handleDeleteAccount(account.id, e)}
                          className={`${selectedAccount?.id === account.id ? 'text-gray-300 hover:text-red-300' : 'text-gray-400 hover:text-red-500'} transition-colors`}
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
                    className="w-full p-3 bg-gray-50 hover:bg-gray-100 rounded-lg text-gray-600 hover:text-gray-700 transition-colors"
                  >
                    + Connect Another Account
                  </button>
                </div>
              )}
            </div>

            {/* Categories Section */}
            <div className="bg-white rounded-lg p-6 shadow-sm">
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
                <div className="mb-4 p-4 rounded-lg bg-gray-50">
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
                    All Emails
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
                        className="w-full text-left p-3 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-800 transition-colors flex justify-between items-center"
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
                        <div className="mt-2 ml-4 space-y-1 border-l-2 border-gray-300 pl-3">
                          {getCustomCategories().map((category) => (
                            <div
                              key={category.id}
                              onClick={() => {
                                handleCategorySelect(category);
                              }}
                              className={`flex justify-between items-center p-2 rounded-lg cursor-pointer transition-colors ${
                                selectedCategory?.id === category.id
                                  ? 'bg-gray-800 text-white'
                                  : 'hover:bg-gray-100 text-gray-700'
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
              <div className="bg-white rounded-lg shadow-sm">
                <div className="px-6 py-4 bg-gray-50 rounded-t-lg">
                  <div className="flex justify-between items-center">
                    <div>
                      <h2 className="text-lg font-medium text-black">
                        {selectedCategory 
                          ? selectedCategory.name
                          : 'All Emails'
                        }
                      </h2>
                      {totalCount > 0 && (
                        <p className="text-sm text-gray-600">
                          {totalCount} {totalCount === 1 ? 'email' : 'emails'}
                        </p>
                      )}
                    </div>
                    <div className="flex gap-2">
                      {selectedEmailIds.size > 0 && (
                        <div className="flex items-center gap-2 bg-blue-50 px-3 py-2 rounded-md">
                          <span className="text-sm text-blue-700 font-medium">
                            {selectedEmailIds.size} selected
                          </span>
                          <button
                            onClick={handleBulkSummarize}
                            disabled={isSummarizing}
                            className="px-3 py-1 text-sm bg-black text-white rounded-md hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                          >
                            {isSummarizing ? (
                              <div className="flex items-center">
                                <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></div>
                                Summarizing...
                              </div>
                            ) : (
                              'AI Summary'
                            )}
                          </button>
                          <button
                            onClick={handleBulkCategorize}
                            disabled={isCategorizing}
                            className="px-3 py-1 text-sm bg-gray-700 text-white rounded-md hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                          >
                            {isCategorizing ? (
                              <div className="flex items-center">
                                <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></div>
                                Categorizing...
                              </div>
                            ) : (
                              'AI Categorize'
                            )}
                          </button>
                          <button
                            onClick={handleBulkUnsubscribe}
                            disabled={isUnsubscribing}
                            className="px-3 py-1 text-sm bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                          >
                            {isUnsubscribing ? (
                              <div className="flex items-center">
                                <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></div>
                                Unsubscribing...
                              </div>
                            ) : (
                              'Unsubscribe'
                            )}
                          </button>
                          <button
                            onClick={() => setSelectedEmailIds(new Set())}
                            className="px-3 py-1 text-sm bg-gray-400 text-white rounded-md hover:bg-gray-500 transition-colors"
                          >
                            Clear
                          </button>
                        </div>
                      )}
                      <button
                        onClick={() => setShowAISummary(!showAISummary)}
                        className={`px-4 py-2 text-sm rounded-md transition-colors ${
                          showAISummary 
                            ? 'bg-gray-800 text-white hover:bg-gray-700' 
                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                        }`}
                      >
                        {showAISummary ? 'Show Email Body' : 'Show AI Summary'}
                      </button>
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
                    <div className="flex justify-between items-start mb-6 gap-4">
                      <button
                        onClick={() => setSelectedEmail(null)}
                        className="inline-flex items-center gap-2 px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 hover:text-gray-900 rounded-lg transition-colors font-medium flex-shrink-0"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                        </svg>
                        Back
                      </button>
                      
                      <div className="flex gap-2 flex-shrink-0">
                        <button
                          onClick={() => handleCategorizeWithAI(selectedEmail.id)}
                          className="px-3 py-2 text-sm bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium whitespace-nowrap"
                        >
                          🤖 Categorize with AI
                        </button>
                        {!selectedEmail.ai_summary && (
                          <button
                            onClick={() => handleGenerateAISummary(selectedEmail.id)}
                            className="px-3 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium whitespace-nowrap"
                          >
                            ✨ Generate Summary
                          </button>
                        )}
                      </div>
                    </div>
                    <div className="space-y-6">
                      {/* Email Header with clear visual separation */}
                      <div className="bg-gray-50 -mx-6 -mt-6 px-6 pt-6 pb-4 border-b border-gray-200">
                        <div className="space-y-4">
                          {/* Subject line with proper overflow handling */}
                          <div className="pr-4">
                            <h1 className="text-2xl font-bold text-black leading-tight break-words">{selectedEmail.subject || '(No Subject)'}</h1>
                          </div>
                          
                          {/* Categories */}
                          {selectedEmail.category_ids && selectedEmail.category_ids.length > 0 && getNonSystemCategoryNames(selectedEmail.category_ids).length > 0 && (
                            <div className="flex flex-wrap gap-1">
                              {getNonSystemCategoryNames(selectedEmail.category_ids).map((categoryName, index) => (
                                <span key={index} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                                  {categoryName}
                                </span>
                              ))}
                            </div>
                          )}
                          
                          {/* Sender and date section with clear visual separation */}
                          <div className="flex items-center justify-between pt-2 border-t border-gray-300">
                            <div>
                              <p className="text-base text-gray-900 font-semibold">{selectedEmail.sender}</p>
                              <p className="text-sm text-gray-600">
                                {new Date(selectedEmail.received_at).toLocaleString('en-US', {
                                  weekday: 'long',
                                  year: 'numeric',
                                  month: 'long',
                                  day: 'numeric',
                                  hour: 'numeric',
                                  minute: '2-digit',
                                  hour12: true
                                })}
                              </p>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      {/* Email Content with clear separation */}
                      <div className="mt-6">
                        <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
                          {showAISummary ? (
                            selectedEmail.ai_summary ? (
                              <div className="bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 rounded-xl p-6 border border-blue-100 shadow-sm">
                                <div className="flex items-start gap-4 mb-4">
                                  <div className="flex-shrink-0">
                                    <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-indigo-600 rounded-lg flex items-center justify-center">
                                      <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                                        <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                      </svg>
                                    </div>
                                  </div>
                                  <div className="flex-1">
                                    <h3 className="text-sm font-semibold text-gray-900 mb-1">AI Summary</h3>
                                    <p className="text-xs text-gray-600">Generated by artificial intelligence</p>
                                  </div>
                                </div>
                                <div className="bg-white/70 rounded-lg p-4 backdrop-blur-sm">
                                  <p className="text-sm text-gray-800 leading-relaxed font-medium">
                                    {selectedEmail.ai_summary}
                                  </p>
                                </div>
                              </div>
                            ) : (
                              <div className="text-center py-8">
                                <p className="text-gray-500 mb-4">No AI summary available</p>
                                <button
                                  onClick={() => handleGenerateAISummary(selectedEmail.id)}
                                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                                >
                                  Generate AI Summary
                                </button>
                              </div>
                            )
                          ) : (
                            selectedEmail.body ? (
                              <div 
                                className="text-sm text-gray-800 leading-relaxed"
                                dangerouslySetInnerHTML={{ 
                                  __html: sanitizeAndRenderEmailBody(selectedEmail.body)
                                }}
                              />
                            ) : (
                              <p className="text-sm text-gray-500 italic">No content available</p>
                            )
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-1">
                    {/* Select All Header */}
                    {emails.length > 0 && (
                      <div className="px-6 py-3 bg-gray-50 border-b border-gray-200">
                        <div className="flex items-center">
                          <input
                            type="checkbox"
                            checked={selectedEmailIds.size === emails.length && emails.length > 0}
                            onChange={handleSelectAll}
                            className="w-5 h-5 rounded-full border-2 border-gray-300 text-black bg-white checked:bg-black checked:border-black focus:ring-2 focus:ring-gray-300 focus:ring-offset-0 transition-all duration-200 ease-in-out hover:border-gray-400 checked:hover:bg-gray-800"
                          />
                        </div>
                      </div>
                    )}
                    {emails.map((email) => (
                      <div
                        key={email.id}
                        className="group px-6 py-4 hover:bg-gray-50 transition-all duration-200 ease-in-out"
                      >
                        <div className="flex items-center gap-4">
                          <div className="flex items-center justify-center">
                            <input
                              type="checkbox"
                              checked={selectedEmailIds.has(email.id)}
                              onChange={(e) => {
                                e.stopPropagation();
                                handleSelectEmail(email.id);
                              }}
                              className="w-5 h-5 rounded-full border-2 border-gray-300 text-black bg-white checked:bg-black checked:border-black focus:ring-2 focus:ring-gray-300 focus:ring-offset-0 transition-all duration-200 ease-in-out hover:border-gray-400 checked:hover:bg-gray-800"
                            />
                          </div>
                          <div 
                            className="flex-1 min-w-0 space-y-1 cursor-pointer"
                            onClick={() => setSelectedEmail(email)}
                          >
                            <div className="flex items-center gap-3">
                              <div className="flex items-center gap-2">
                                <h3 className="font-semibold text-black text-base truncate group-hover:text-gray-900">
                                  {email.subject || '(No Subject)'}
                                </h3>
                                {email.unsubscribe_link && (
                                  <span 
                                    className="inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-600 border border-red-200"
                                    title="Unsubscribe available"
                                  >
                                    📧
                                  </span>
                                )}
                              </div>
                              {email.category_ids && email.category_ids.length > 0 && getNonSystemCategoryNames(email.category_ids).length > 0 && (
                                <div className="flex flex-wrap gap-1">
                                  {getNonSystemCategoryNames(email.category_ids).slice(0, 2).map((categoryName, index) => (
                                    <span key={index} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-700 border-0">
                                      {categoryName}
                                    </span>
                                  ))}
                                  {getNonSystemCategoryNames(email.category_ids).length > 2 && (
                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-200 text-gray-600 border-0">
                                      +{getNonSystemCategoryNames(email.category_ids).length - 2}
                                    </span>
                                  )}
                                </div>
                              )}
                            </div>
                            <div className="flex items-center justify-between">
                              <p className="text-sm text-gray-600 truncate font-medium">
                                {email.sender}
                              </p>
                              <p className="text-xs text-gray-500 font-normal">
                                {new Date(email.received_at).toLocaleString('en-US', {
                                  month: 'short',
                                  day: 'numeric',
                                  year: new Date(email.received_at).getFullYear() !== new Date().getFullYear() ? 'numeric' : undefined,
                                  hour: 'numeric',
                                  minute: '2-digit',
                                  hour12: true
                                })}
                              </p>
                            </div>
                            {(showAISummary || emailsShowingAISummary.has(email.id)) ? (
                              email.ai_summary && (
                                <div className="mt-2 p-2 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border-l-3 border-blue-400">
                                  <div className="flex items-start gap-2">
                                    <div className="flex-shrink-0 mt-0.5">
                                      <div className="w-1.5 h-1.5 bg-blue-500 rounded-full"></div>
                                    </div>
                                    <p className="text-xs text-blue-800 font-medium leading-relaxed">
                                      {email.ai_summary}
                                    </p>
                                  </div>
                                </div>
                              )
                            ) : (
                              email.body && (
                                <div 
                                  className="text-xs text-gray-700 font-normal truncate"
                                  dangerouslySetInnerHTML={{ 
                                    __html: sanitizeAndRenderEmailBody(email.body).substring(0, 200) + '...'
                                  }}
                                />
                              )
                            )}
                          </div>
                          <div className="flex-shrink-0 flex items-center gap-2">
                            <div className="opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex gap-1">
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleCategorizeWithAI(email.id);
                                }}
                                className="p-1 text-xs bg-purple-100 text-purple-600 rounded hover:bg-purple-200 transition-colors"
                                title="Categorize with AI"
                              >
                                🤖
                              </button>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleToggleEmailAISummary(email.id, email);
                                }}
                                className={`p-1 text-xs rounded transition-colors ${
                                  email.ai_summary 
                                    ? (emailsShowingAISummary.has(email.id) && !showAISummary)
                                      ? 'bg-blue-100 text-blue-600 hover:bg-blue-200'
                                      : 'bg-green-100 text-green-600 hover:bg-green-200'
                                    : 'bg-blue-100 text-blue-600 hover:bg-blue-200'
                                }`}
                                title={
                                  email.ai_summary 
                                    ? (emailsShowingAISummary.has(email.id) && !showAISummary)
                                      ? "Hide AI Summary"
                                      : "Show AI Summary"
                                    : "Generate AI Summary"
                                }
                              >
                                {email.ai_summary 
                                  ? (emailsShowingAISummary.has(email.id) && !showAISummary)
                                    ? '🔽'
                                    : '✅'
                                  : '✨'
                                }
                              </button>
                              {email.unsubscribe_link && (
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleSingleUnsubscribe(email.id);
                                  }}
                                  className="p-1 text-xs bg-red-100 text-red-600 rounded hover:bg-red-200 transition-colors"
                                  title="Unsubscribe"
                                >
                                  📧
                                </button>
                              )}
                            </div>
                            <div className="w-2 h-2 rounded-full bg-gray-300 opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                
                {/* Pagination */}
                {!emailsLoading && !selectedEmail && emails.length > 0 && totalPages > 1 && (
                  <div className="px-6 py-6 bg-gray-50">
                    <div className="flex items-center justify-between">
                      <div className="text-sm text-gray-600 font-medium">
                        Page {currentPage} of {totalPages}
                      </div>
                      <div className="flex space-x-1">
                        <button
                          onClick={() => handlePageChange(currentPage - 1)}
                          disabled={currentPage <= 1}
                          className="px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors"
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
                              className={`px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                                pageNum === currentPage
                                  ? 'bg-black text-white'
                                  : 'bg-white text-gray-700 hover:bg-gray-100'
                              }`}
                            >
                              {pageNum}
                            </button>
                          );
                        })}
                        
                        <button
                          onClick={() => handlePageChange(currentPage + 1)}
                          disabled={currentPage >= totalPages}
                          className="px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors"
                        >
                          Next
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="bg-white rounded-lg shadow-sm p-6">
                <div className="text-center py-12">
                  <p className="text-gray-500">Select a Gmail account to view emails</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
    </>
  );
}