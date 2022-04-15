package pre_migrate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/nats-v2-migrate/fakes"
	"code.cloudfoundry.org/nats-v2-migrate/nats"
	pre_migrate "code.cloudfoundry.org/nats-v2-migrate/pre-migrate"
)

var _ = Describe("PreMigrator", func() {
	var (
		natsConn1     *fakes.NatsConn
		natsConn2     *fakes.NatsConn
		natsConn3     *fakes.NatsConn
		natsConns     []nats.NatsConn
		rewriter      *fakes.Rewriter
		natsBPMPath   string
		natsV1BPMPath string
		pre_migrator  *pre_migrate.PreMigrator
	)
	BeforeEach(func() {
		natsBPMPath = "/var/vcap/jobs/nats-tls/config/bpm.yml"
		natsV1BPMPath = "/var/vcap/jobs/nats-tls/config/bpm.v1.yml"
		natsConn1 = &fakes.NatsConn{}
		natsConn2 = &fakes.NatsConn{}
		natsConn3 = &fakes.NatsConn{}
		rewriter = &fakes.Rewriter{}
	})

	Describe("NewPreMigrator", func() {
		It("Creates the PreMigrator object with expeted properties", func() {

			natsConns = []nats.NatsConn{natsConn1, natsConn2, natsConn3}
			rewriter = &fakes.Rewriter{}
			pre_migrator = pre_migrate.NewPreMigrator(natsConns, rewriter, natsV1BPMPath, natsBPMPath)
			Expect(pre_migrator).To(Equal(&pre_migrate.PreMigrator{
				NatsConns:     natsConns,
				BpmRewriter:   rewriter,
				NatsV1BpmPath: natsV1BPMPath,
				NatsBpmPath:   natsBPMPath,
			}))
		})
	})

	Describe("PrepareForMigration", func() {
		JustBeforeEach(func() {
			natsConns = []nats.NatsConn{natsConn1, natsConn2, natsConn3}
			pre_migrator = pre_migrate.NewPreMigrator(natsConns, rewriter, natsV1BPMPath, natsBPMPath)
		})

		Context("There are nats v1 machines in the cluster", func() {
			BeforeEach(func() {
				natsConn1.ConnectedServerVersionReturns("1.1.1")
			})

			It("does replace the bpm config", func() {
				err := pre_migrator.PrepareForMigration()
				Expect(err).NotTo(HaveOccurred())

				Expect(natsConn1.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn2.ConnectedServerVersionCallCount()).To(Equal(0))
				Expect(natsConn3.ConnectedServerVersionCallCount()).To(Equal(0))

				Expect(rewriter.RewriteCallCount()).To(Equal(1))
				arg1, arg2 := rewriter.RewriteArgsForCall(0)
				Expect(arg1).To(Equal("/var/vcap/jobs/nats-tls/config/bpm.yml"))
				Expect(arg2).To(Equal("/var/vcap/jobs/nats-tls/config/bpm.v1.yml"))
			})
		})

		Context("There are some nats v1 and some v2 machines in the cluster", func() {
			BeforeEach(func() {
				natsConn1.ConnectedServerVersionReturns("2.2.2")
				natsConn2.ConnectedServerVersionReturns("1.1.1")
			})

			It("does replace the bpm config", func() {
				err := pre_migrator.PrepareForMigration()
				Expect(err).NotTo(HaveOccurred())

				Expect(natsConn1.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn2.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn3.ConnectedServerVersionCallCount()).To(Equal(0))

				Expect(rewriter.RewriteCallCount()).To(Equal(1))
				arg1, arg2 := rewriter.RewriteArgsForCall(0)
				Expect(arg1).To(Equal("/var/vcap/jobs/nats-tls/config/bpm.yml"))
				Expect(arg2).To(Equal("/var/vcap/jobs/nats-tls/config/bpm.v1.yml"))
			})
		})

		Context("There are only v2 machines in the cluster", func() {
			BeforeEach(func() {
				natsConn1.ConnectedServerVersionReturns("2.2.2")
				natsConn2.ConnectedServerVersionReturns("2.2.2")
				natsConn3.ConnectedServerVersionReturns("2.2.2")
			})

			It("does replace the bpm config", func() {
				err := pre_migrator.PrepareForMigration()
				Expect(err).NotTo(HaveOccurred())

				Expect(natsConn1.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn2.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn3.ConnectedServerVersionCallCount()).To(Equal(1))

				Expect(rewriter.RewriteCallCount()).To(Equal(0))
			})
		})
		Context("Unexpected semantnic version format", func() {
			BeforeEach(func() {
				natsConn1.ConnectedServerVersionReturns("1.0")
			})
			It("does not replace the bpm config", func() {

				err := pre_migrator.PrepareForMigration()
				Expect(err).To(HaveOccurred())

				Expect(natsConn1.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn2.ConnectedServerVersionCallCount()).To(Equal(0))
				Expect(natsConn3.ConnectedServerVersionCallCount()).To(Equal(0))

				Expect(rewriter.RewriteCallCount()).To(Equal(0))
			})
		})

		Context("Invalid version response", func() {
			BeforeEach(func() {
				natsConn1.ConnectedServerVersionReturns("notanumber")
			})
			It("does not replace the bpm config", func() {

				err := pre_migrator.PrepareForMigration()
				Expect(err).To(HaveOccurred())

				Expect(natsConn1.ConnectedServerVersionCallCount()).To(Equal(1))
				Expect(natsConn2.ConnectedServerVersionCallCount()).To(Equal(0))
				Expect(natsConn3.ConnectedServerVersionCallCount()).To(Equal(0))

				Expect(rewriter.RewriteCallCount()).To(Equal(0))
			})
		})
	})

})
