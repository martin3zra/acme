import { PageProps } from '@/types';
import { Link } from '@inertiajs/react';
import { Box, Calendar, CheckCircle, CreditCard, FileText, Layers, Users } from 'lucide-react';

export default function Index({ auth }: PageProps) {
  // --- Data & small helpers
  const FeatureCard = ({ icon: Icon, title, children }: { icon: React.ElementType; title: string; children: React.ReactNode }) => (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-800">
      <div className="flex items-center gap-4">
        <div className="rounded-lg bg-indigo-50 p-3 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-300">
          <Icon className="h-6 w-6" />
        </div>
        <div>
          <h3 className="text-lg font-semibold">{title}</h3>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-300">{children}</p>
        </div>
      </div>
    </div>
  );
  const Stat = ({ value, label }: { value: string; label: string }) => (
    <div className="text-center">
      <div className="text-2xl font-bold">{value}</div>
      <div className="text-sm text-gray-500">{label}</div>
    </div>
  );

  // --- Components
  const Hero = () => (
    <section className="bg-gradient-to-b from-white to-gray-50 py-20 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-1 items-center gap-12 md:grid-cols-2">
          <div>
            <h1 className="text-4xl leading-tight font-extrabold sm:text-5xl">
              Run your business smarter — invoicing, inventory & payments in one place
            </h1>
            <p className="mt-4 max-w-2xl text-lg text-gray-600 dark:text-gray-300">
              Manage invoices, stock, payments, customers and team members across multiple businesses — from a single unified dashboard. Custom
              sequences, PDF invoices, advanced reporting and roles & permissions out of the box.
            </p>

            <div className="mt-8 flex flex-wrap gap-3">
              <a
                className="inline-flex items-center rounded-lg bg-indigo-600 px-5 py-3 font-semibold text-white shadow hover:bg-indigo-700"
                href="#signup"
              >
                Start Free Trial
              </a>
              <a
                className="inline-flex items-center rounded-lg border border-gray-200 px-5 py-3 text-gray-700 dark:border-gray-700 dark:text-gray-200"
                href="#demo"
              >
                Live Demo
              </a>
            </div>

            <div className="mt-8 grid max-w-sm grid-cols-3 gap-4">
              <Stat value="50k+" label="Invoices processed" />
              <Stat value="10k+" label="Products tracked" />
              <Stat value="1k+" label="Businesses" />
            </div>
          </div>

          <div className="relative">
            {/* dashboard mock */}
            <div className="overflow-hidden rounded-2xl border border-gray-100 shadow-2xl dark:border-gray-700">
              <div className="bg-gradient-to-br from-indigo-50 to-white p-6 dark:from-gray-800 dark:to-gray-900">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm text-gray-500">Business: Acme Corp</div>
                    <div className="mt-2 text-2xl font-semibold">Overview</div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-500">Balance</div>
                    <div className="text-xl font-bold">$24,360.34</div>
                  </div>
                </div>

                <div className="mt-6 grid grid-cols-2 gap-4">
                  <div className="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800">
                    <div className="text-xs text-gray-500">Open Invoices</div>
                    <div className="mt-2 font-semibold">12</div>
                  </div>
                  <div className="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800">
                    <div className="text-xs text-gray-500">Due This Month</div>
                    <div className="mt-2 font-semibold">$6,420</div>
                  </div>
                </div>

                <div className="mt-6 rounded-lg border border-gray-100 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
                  <div className="text-xs text-gray-500">Recent Activity</div>
                  <ul className="mt-2 space-y-2 text-sm">
                    <li>Invoice INV-2025-00123 paid — $1,200</li>
                    <li>Stock updated — Product: USB-C Cable</li>
                    <li>New user added — jane@acme.com</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
  const Features = () => (
    <section id="features" className="py-20">
      <div className="container mx-auto px-4">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-3xl font-bold">Everything you need — billing, inventory, payments, teams</h2>
          <p className="mt-3 text-gray-600 dark:text-gray-300">
            Built for growing businesses who need accurate billing, clean reporting and a simple way to manage multiple teams and businesses.
          </p>
        </div>

        <div className="mt-10 grid grid-cols-1 gap-6 md:grid-cols-3">
          <FeatureCard icon={FileText} title="Smart Invoicing">
            Custom templates, automatic PDFs, sequence formats, recurring invoices.
          </FeatureCard>
          <FeatureCard icon={Box} title="Inventory & Items">
            Real-time stock, SKUs, CSV import/export, low-stock alerts.
          </FeatureCard>
          <FeatureCard icon={CreditCard} title="Payments">
            Track payments, partial payments, reconcile and export transactions.
          </FeatureCard>
        </div>

        <div className="mt-6 grid grid-cols-1 gap-6 md:grid-cols-3">
          <FeatureCard icon={Users} title="Multi-User & Roles">
            Invite teammates, role-based permissions and audit logs.
          </FeatureCard>
          <FeatureCard icon={Layers} title="Multi-Business">
            Manage multiple companies, switch contexts, separate settings and sequences.
          </FeatureCard>
          <FeatureCard icon={Calendar} title="Reporting & Schedules">
            Scheduled reports, DSL-based custom queries and exports.
          </FeatureCard>
        </div>
      </div>
    </section>
  );
  const Workflow = () => (
    <section className="bg-gray-50 py-16 dark:bg-gray-900">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-1 items-start gap-6 md:grid-cols-3">
          <div className="md:col-span-1">
            <h3 className="text-2xl font-bold">How it works</h3>
            <p className="mt-3 text-gray-600 dark:text-gray-300">Get set up in minutes and manage everything from a single dashboard.</p>
          </div>
          <div className="md:col-span-2">
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
              <div className="rounded-xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-md bg-indigo-100 p-2 text-indigo-600 dark:bg-indigo-900/30">
                    <CheckCircle className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="font-semibold">Create your business</h4>
                    <p className="text-sm text-gray-500">Add company details, currency and invoice format.</p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-md bg-indigo-100 p-2 text-indigo-600 dark:bg-indigo-900/30">
                    <Box className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="font-semibold">Add customers & items</h4>
                    <p className="text-sm text-gray-500">Import via CSV or add manually — track SKUs and stock.</p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-md bg-indigo-100 p-2 text-indigo-600 dark:bg-indigo-900/30">
                    <CreditCard className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="font-semibold">Send invoices & collect payments</h4>
                    <p className="text-sm text-gray-500">Generate branded PDFs and track payments instantly.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
  const Testimonials = () => (
    <section className="py-16">
      <div className="container mx-auto px-4">
        <div className="text-center">
          <h3 className="text-2xl font-bold">Loved by businesses everywhere</h3>
          <p className="mt-3 text-gray-600 dark:text-gray-300">Real results from real customers.</p>
        </div>

        <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
          <blockquote className="rounded-2xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30">JD</div>
              <div>
                <div className="font-semibold">Jorge D. — Bakery Co.</div>
                <div className="text-sm text-gray-500">"Saved 12 hrs/week on billing and our stock finally matches reality."</div>
              </div>
            </div>
          </blockquote>

          <blockquote className="rounded-2xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30">EM</div>
              <div>
                <div className="font-semibold">Elena M. — Design Studio</div>
                <div className="text-sm text-gray-500">"Invoices, payments and reports in a single product. Game changer."</div>
              </div>
            </div>
          </blockquote>

          <blockquote className="rounded-2xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30">AR</div>
              <div>
                <div className="font-semibold">Andrés R. — Electronics</div>
                <div className="text-sm text-gray-500">"Multi-business support made it easy to manage my stores from one account."</div>
              </div>
            </div>
          </blockquote>
        </div>
      </div>
    </section>
  );

  const Pricing = () => (
    <section id="pricing" className="bg-white py-16 dark:bg-gray-900">
      <div className="container mx-auto px-4">
        <div className="mx-auto max-w-3xl text-center">
          <h3 className="text-2xl font-bold">Simple pricing that scales</h3>
          <p className="mt-2 text-gray-600 dark:text-gray-300">Start free and upgrade as your needs grow.</p>
        </div>

        <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
          <div className="rounded-2xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="text-lg font-semibold">Starter</div>
            <div className="mt-2 text-3xl font-bold">
              $19<span className="text-sm font-medium">/mo</span>
            </div>
            <ul className="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
              <li>1 business</li>
              <li>Unlimited invoices</li>
              <li>Inventory basics</li>
            </ul>
            <a href="#signup" className="mt-6 inline-block w-full rounded-lg bg-indigo-600 px-4 py-2 text-center font-semibold text-white">
              Start
            </a>
          </div>

          <div className="rounded-2xl border border-indigo-100 bg-indigo-50 p-6 dark:border-indigo-900/30 dark:bg-indigo-900/10">
            <div className="text-lg font-semibold">Business</div>
            <div className="mt-2 text-3xl font-bold">
              $49<span className="text-sm font-medium">/mo</span>
            </div>
            <ul className="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
              <li>Unlimited businesses</li>
              <li>5 users</li>
              <li>Advanced reports</li>
            </ul>
            <a href="#signup" className="mt-6 inline-block w-full rounded-lg bg-indigo-600 px-4 py-2 text-center font-semibold text-white">
              Choose
            </a>
          </div>

          <div className="rounded-2xl border border-gray-100 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="text-lg font-semibold">Enterprise</div>
            <div className="mt-2 text-3xl font-bold">Custom</div>
            <ul className="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
              <li>Priority support</li>
              <li>SSO & integrations</li>
              <li>Custom SLAs</li>
            </ul>
            <a href="#contact" className="mt-6 inline-block w-full rounded-lg border border-gray-200 px-4 py-2 text-center dark:border-gray-700">
              Contact
            </a>
          </div>
        </div>
      </div>
    </section>
  );
  const FAQ = () => (
    <section id="faq" className="bg-gray-50 py-16 dark:bg-gray-900">
      <div className="container mx-auto px-4">
        <div className="mx-auto max-w-3xl text-center">
          <h3 className="text-2xl font-bold">Frequently asked questions</h3>
          <div className="mt-6 space-y-4 text-left">
            <details className="rounded-xl border border-gray-100 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
              <summary className="font-semibold">Can I manage multiple businesses?</summary>
              <div className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                Yes — accounts can contain multiple businesses, each with separate settings and sequences.
              </div>
            </details>

            <details className="rounded-xl border border-gray-100 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
              <summary className="font-semibold">Do you support CSV import/export?</summary>
              <div className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                Yes — customers, items and invoices can be imported or exported via CSV.
              </div>
            </details>

            <details className="rounded-xl border border-gray-100 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
              <summary className="font-semibold">Is my data secure?</summary>
              <div className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                We use industry standard encryption in transit and at rest. Enterprise plans offer additional controls, audits, and SSO.
              </div>
            </details>
          </div>
        </div>
      </div>
    </section>
  );

  const Footer = () => (
    <footer className="border-t border-gray-100 py-8 dark:border-gray-800">
      <div className="container mx-auto flex flex-col items-center justify-between gap-4 px-4 md:flex-row">
        <div className="flex items-center gap-3">
          <div className="text-2xl font-bold">Acme</div>
          <div className="text-sm text-gray-500">— Invoicing & Inventory</div>
        </div>
        <div className="text-sm text-gray-500">© {new Date().getFullYear()} Acme. All rights reserved.</div>
      </div>
    </footer>
  );
  return (
    <div className="min-h-screen bg-white font-sans text-gray-900 dark:bg-black dark:text-gray-100">
      <header className="border-b border-gray-100 py-6 dark:border-gray-800">
        <div className="container mx-auto flex items-center justify-between px-4">
          <div className="flex items-center gap-3">
            <div className="text-xl font-bold">Acme</div>
            <div className="hidden text-sm text-gray-500 sm:block">Invoicing • Inventory • Payments</div>
          </div>
          <nav className="flex items-center gap-4">
            <a href="#features" className="text-sm text-gray-700 dark:text-gray-300">
              Features
            </a>
            <a href="#pricing" className="text-sm text-gray-700 dark:text-gray-300">
              Pricing
            </a>
            <a href="#faq" className="text-sm text-gray-700 dark:text-gray-300">
              FAQ
            </a>
            {auth.user ? (
              <Link className="rounded-lg bg-indigo-600 px-4 py-2 text-sm text-white" href="/home">
                Dashboard
              </Link>
            ) : (
              <Link href="/login" className="rounded-lg bg-indigo-600 px-4 py-2 text-sm text-white">
                Book a Demo
              </Link>
            )}
          </nav>
        </div>
      </header>

      <main>
        <Hero />
        <Features />
        <Workflow />
        <Testimonials />
        <Pricing />
        <FAQ />
      </main>

      <Footer />
    </div>
  );
}
