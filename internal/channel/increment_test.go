package channel

import (
	"context"

	"github.com/Pallinder/go-randomdata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/channel"
)

var _ = Describe("Incrementing subscribers for a channel @increment", func() {
	var (
		incReq *channel.SubscribersRequest
		ctx    context.Context
	)

	BeforeEach(func() {
		incReq = &channel.SubscribersRequest{
			Id: randomdata.RandStringRunes(20),
		}
		ctx = context.Background()
	})

	Describe("Incrementing with malformed request", func() {
		It("should fail when the request is nil", func() {
			incReq = nil
			incRes, err := ChannelAPI.IncrementSubscribers(ctx, incReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(incRes).Should(BeNil())
		})
		It("should fail when channel id is missing", func() {
			incReq.Id = ""
			incRes, err := ChannelAPI.IncrementSubscribers(ctx, incReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(incRes).Should(BeNil())
		})
		It("should fail when channel id is incorrect", func() {
			incReq.Id = randomdata.RandStringRunes(32)
			incRes, err := ChannelAPI.IncrementSubscribers(ctx, incReq)
			Expect(err).Should(HaveOccurred())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			Expect(incRes).Should(BeNil())
		})
	})

	Describe("Incrementing channel with well-formed request", func() {
		var channelID string
		Context("Lets create one channel first", func() {
			It("should succeed", func() {
				createReq := &channel.CreateChannelRequest{
					Channel: mockChannel(),
				}
				createRes, err := ChannelAPI.CreateChannel(ctx, createReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(createRes).ShouldNot(BeNil())
				channelID = createRes.Id
			})
		})

		count := 20

		When("Getting the channel", func() {
			It("should succeed", func() {
				getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
					Id: channelID,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status.Code(err)).Should(Equal(codes.OK))
				Expect(getRes).ShouldNot(BeNil())
			})

			for i := 0; i < count; i++ {
				It("should succeed", func() {
					incReq.Id = channelID
					incRes, err := ChannelAPI.IncrementSubscribers(ctx, incReq)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(incRes).ShouldNot(BeNil())
				})
			}

			When("Getting the channel", func() {
				It("should succeed", func() {
					getRes, err := ChannelAPI.GetChannel(ctx, &channel.GetChannelRequest{
						Id: channelID,
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(status.Code(err)).Should(Equal(codes.OK))
					Expect(getRes).ShouldNot(BeNil())
					Expect(getRes.Subscribers).Should(BeEquivalentTo(count))
				})
			})
		})
	})
})
